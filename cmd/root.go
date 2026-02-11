package cmd

import (
	"fmt"
	"os"
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
	var mu sync.Mutex

	fmt.Println()
	headerColor.Printf("%-40s | %-15s | %-10s | %-15s | %-20s\n", "Account (Email)", "Provider", "Remaining", "Reset In", "Model")
	headerColor.Println(strings.Repeat("-", 115))

	for _, file := range files {
		if file.Disabled {
			continue
		}

		wg.Add(1)
		go func(f models.AuthFile) {
			defer wg.Done()
			limits, err := client.FetchQuota(f)
			if err != nil {
				return
			}

			for modelName, limit := range limits {
				displayModelName := utils.GetDisplayModelName(modelName, f.Provider, fullMode)
				if displayModelName == "" {
					continue
				}

				remainingStr := strings.TrimSuffix(limit.Remaining, "%")
				remainingVal, _ := strconv.Atoi(remainingStr)
				quotaColor := utils.GetQuotaColor(remainingVal)
				resetStr := utils.GetResetString(limit.ResetTime)

				mu.Lock()
				emailColor.Printf("%-40s | ", f.Email)
				fmt.Printf("%-15s | ", f.Provider)
				quotaColor.Printf("%-10s | ", limit.Remaining)
				fmt.Printf("%-15s | %-20s\n", resetStr, displayModelName)
				mu.Unlock()
			}
		}(file)
	}
	wg.Wait()
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
