package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/quaywin/quota-sense-cli/internal/api"
	"github.com/quaywin/quota-sense-cli/internal/config"
	"github.com/quaywin/quota-sense-cli/internal/models"
	"github.com/quaywin/quota-sense-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	headerColor  = color.New(color.FgCyan, color.Bold)
	emailColor   = color.New(color.FgWhite)
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	fullMode     bool
)

var rootCmd = &cobra.Command{
	Use:   "qs",
	Short: "QuotaSense CLI - Monitor your AI model usage",
	Long:  `QuotaSense is a CLI tool to monitor and manage your AI model usage quotas from the terminal.`,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "update" || cmd.Name() == "version" {
			return
		}
		checkAndNotifyUpdate()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			cfg, err = config.PromptConfig()
			if err != nil {
				errorColor.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			client := api.NewClient(cfg)
			fmt.Println("Verifying connection...")
			if err := client.CheckConnection(); err != nil {
				errorColor.Printf("Connection failed: %v\n", err)
				os.Exit(1)
			}

			if err := config.SaveConfig(cfg); err != nil {
				errorColor.Printf("Error saving config: %v\n", err)
				os.Exit(1)
			}
			successColor.Println("Configuration saved successfully!")
		}

		displayQuota(cfg)
	},
}

type displayEntry struct {
	limit            models.ModelLimit
	displayModelName string
}

type accountResult struct {
	file        models.AuthFile
	err         error
	bestInGroup map[string]displayEntry
}

func displayQuota(cfg *config.Config) {
	if cfg == nil {
		return
	}
	client := api.NewClient(cfg)
	fmt.Println("Fetching usage information...")

	files, err := client.FetchUsage()
	if err != nil {
		errorColor.Printf("Error fetching usage: %v\n", err)
		return
	}

	var wg sync.WaitGroup

	fmt.Println()
	if fullMode {
		headerColor.Printf("%-40s | %-15s | %-10s | %-15s | %-25s | %-20s\n", "Account (Email)", "Provider", "Remaining", "Reset In", "Model Name", "Model")
		headerColor.Println(strings.Repeat("-", 140))
	} else {
		headerColor.Printf("%-40s | %-15s | %-10s | %-15s | %-20s\n", "Account (Email)", "Provider", "Remaining", "Reset In", "Model")
		headerColor.Println(strings.Repeat("-", 115))
	}

	results := make([]accountResult, len(files))

	for i, file := range files {
		wg.Add(1)
		go func(idx int, f models.AuthFile) {
			defer wg.Done()
			limits, err := client.FetchQuota(f)
			res := accountResult{
				file: f,
				err:  err,
			}
			if err == nil {
				bestInGroup := make(map[string]displayEntry)
				for modelName, limit := range limits {
					displayModelName := utils.GetDisplayModelName(modelName, f.Provider, fullMode)
					if displayModelName == "" {
						continue
					}

					key := displayModelName
					if fullMode {
						key = modelName
					}

					if existing, ok := bestInGroup[key]; !ok || limit.RemainingFraction < existing.limit.RemainingFraction {
						bestInGroup[key] = displayEntry{limit, displayModelName}
					}
				}
				res.bestInGroup = bestInGroup
			}
			results[idx] = res
		}(i, file)
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		return !results[i].file.Disabled && results[j].file.Disabled
	})

	for _, res := range results {
		f := res.file
		err := res.err
		bestInGroup := res.bestInGroup

		if err != nil {
			if f.Disabled {
				disabledColor := color.New(color.FgHiBlack)
				emailStr := f.Email
				if !strings.Contains(emailStr, "(disabled)") {
					emailStr += " (disabled)"
				}
				disabledColor.Printf("%-40s | ", emailStr)
				disabledColor.Printf("%-15s | ", f.Provider)
				disabledColor.Printf("%-10s | ", "Disabled")
				if fullMode {
					disabledColor.Printf("%-15s | %-25s | %-20s\n", "-", "-", "-")
				} else {
					disabledColor.Printf("%-15s | %-20s\n", "-", "-")
				}
			}
			continue
		}

		if len(bestInGroup) == 0 && f.Disabled {
			disabledColor := color.New(color.FgHiBlack)
			emailStr := f.Email
			if !strings.Contains(emailStr, "(disabled)") {
				emailStr += " (disabled)"
			}
			disabledColor.Printf("%-40s | ", emailStr)
			disabledColor.Printf("%-15s | ", f.Provider)
			disabledColor.Printf("%-10s | ", "Disabled")
			if fullMode {
				disabledColor.Printf("%-15s | %-25s | %-20s\n", "-", "-", "-")
			} else {
				disabledColor.Printf("%-15s | %-20s\n", "-", "-")
			}
			continue
		}

		for _, entry := range bestInGroup {
			remainingStr := strings.TrimSuffix(entry.limit.Remaining, "%")
			remainingVal, err := strconv.Atoi(remainingStr)

			var quotaColor *color.Color
			var rowColor *color.Color
			var modelColor *color.Color

			if f.Disabled {
				rowColor = color.New(color.FgHiBlack)
				quotaColor = rowColor
				modelColor = rowColor
			} else {
				rowColor = color.New(color.FgWhite)
				if err == nil {
					quotaColor = utils.GetQuotaColor(remainingVal)
					if remainingVal == 0 {
						modelColor = color.New(color.FgRed, color.Bold)
					} else {
						modelColor = rowColor
					}
				} else {
					quotaColor = color.New(color.FgWhite)
					modelColor = rowColor
				}
			}

			resetStr := utils.GetResetString(entry.limit.ResetTime)

			emailStr := f.Email
			if f.Disabled && !strings.Contains(emailStr, "(disabled)") {
				emailStr += " (disabled)"
			}

			rowColor.Printf("%-40s | ", emailStr)
			rowColor.Printf("%-15s | ", f.Provider)
			quotaColor.Printf("%-10s | ", entry.limit.Remaining)
			if fullMode {
				rowColor.Printf("%-15s | ", resetStr)
				modelColor.Printf("%-25s | ", entry.limit.DisplayName)
				modelColor.Printf("%-20s\n", entry.displayModelName)
			} else {
				rowColor.Printf("%-15s | ", resetStr)
				modelColor.Printf("%-20s\n", entry.displayModelName)
			}
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&fullMode, "full", "f", false, "Display all available models")
}
