package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Review quota and planned usage",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking current quota status...")
		// Future: Add logic to fetch data from API
		fmt.Println("Model: claude-3-5-sonnet | Usage: 45% | Status: Healthy")

		fmt.Print("\nDo you want to proceed? (y/n): ")
		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" {
			fmt.Println("Proceeding with planned action...")
		} else {
			fmt.Println("Action cancelled.")
		}
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
