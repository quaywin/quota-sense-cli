package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/quaywin/quota-sense-cli/internal/api"
	"github.com/quaywin/quota-sense-cli/internal/config"
	"github.com/quaywin/quota-sense-cli/internal/models"
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
	fmt.Println()
	headerColor.Printf("%-40s | %-15s | %-20s | %-10s | %-15s\n", "Account (Email)", "Provider", "Model", "Remaining", "Reset In")
	headerColor.Println(strings.Repeat("-", 115))

	for _, file := range files {
		if file.Disabled || file.Unavailable {
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
				displayModelName := modelName
				if !fullMode {
					if f.Provider == "antigravity" {
						switch modelName {
						case "gemini-3-pro-high":
							displayModelName = "Gemini 3 Pro"
						case "gemini-3-flash":
							displayModelName = "Gemini 3 Flash"
						case "claude-sonnet-4-5":
							displayModelName = "Claude/GPT"
						default:
							continue
						}
					} else if f.Provider == "gemini-cli" {
						switch modelName {
						case "gemini-3-pro-preview":
							displayModelName = "Gemini Pro"
						case "gemini-3-flash-preview":
							displayModelName = "Gemini Flash"
						default:
							continue
						}
					}
				}

				remainingStr := strings.TrimSuffix(limit.Remaining, "%")
				remainingVal, _ := strconv.Atoi(remainingStr)

				var quotaColor *color.Color
				if remainingVal > 50 {
					quotaColor = color.New(color.FgGreen)
				} else if remainingVal > 20 {
					quotaColor = color.New(color.FgYellow)
				} else {
					quotaColor = color.New(color.FgRed, color.Bold)
				}

				resetStr := "-"
				if limit.ResetTime != "" {
					resetTime, err := time.Parse(time.RFC3339, limit.ResetTime)
					if err == nil {
						duration := time.Until(resetTime)
						if duration > 0 {
							resetStr = formatDuration(duration)
						} else {
							resetStr = "Now"
						}
					}
				}

				emailColor.Printf("%-40s | ", f.Email)
				fmt.Printf("%-15s | %-20s | ", f.Provider, displayModelName)
				quotaColor.Printf("%-10s | ", limit.Remaining)
				fmt.Printf("%-15s\n", resetStr)
			}
		}(file)
	}
	wg.Wait()
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
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
