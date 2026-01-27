package cmd

import (
	"fmt"
	"os"

	"github.com/quaywin/quota-sense-cli/internal/api"
	"github.com/quaywin/quota-sense-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure remote server connection",
	Long:  `Set or update the remote server URL and management token for QuotaSense.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.PromptConfig()
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
		successColor.Println("Configuration updated successfully!")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
