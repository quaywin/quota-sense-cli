package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

type UpdateCache struct {
	LastChecked   time.Time `json:"last_checked"`
	LatestVersion string    `json:"latest_version"`
}

func getUpdateCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".quota-sense-update.json")
}

func loadUpdateCache() *UpdateCache {
	path := getUpdateCachePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return &UpdateCache{}
	}

	var cache UpdateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return &UpdateCache{}
	}

	return &cache
}

func saveUpdateCache(cache *UpdateCache) {
	path := getUpdateCachePath()
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0600)
}

func parseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ".")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		// Strip any non-digit suffix (e.g., "-draft" or "-rc1")
		digitStr := ""
		for _, char := range p {
			if char >= '0' && char <= '9' {
				digitStr += string(char)
			} else {
				break
			}
		}
		if digitStr == "" {
			res = append(res, 0)
			continue
		}
		var val int
		fmt.Sscanf(digitStr, "%d", &val)
		res = append(res, val)
	}
	return res
}

func isNewerVersion(current, latest string) bool {
	if current == "dev" || current == "" || latest == "" {
		return false
	}
	if current == latest {
		return false
	}

	currParts := parseVersion(current)
	lateParts := parseVersion(latest)

	// Compare component by component
	maxLen := len(currParts)
	if len(lateParts) > maxLen {
		maxLen = len(lateParts)
	}

	for i := 0; i < maxLen; i++ {
		currVal := 0
		if i < len(currParts) {
			currVal = currParts[i]
		}
		lateVal := 0
		if i < len(lateParts) {
			lateVal = lateParts[i]
		}

		if lateVal > currVal {
			return true
		}
		if currVal > lateVal {
			return false
		}
	}

	return false
}

func checkAndNotifyUpdate() {
	if Version == "dev" {
		return
	}

	cache := loadUpdateCache()
	now := time.Now()

	// If 24 hours have passed since last check, fetch the latest version
	if now.Sub(cache.LastChecked) > 24*time.Hour {
		latest, err := getLatestRelease()
		if err == nil && latest != nil {
			cache.LatestVersion = latest.TagName
		}
		// Always update LastChecked to avoid spamming network requests on failure/offline
		cache.LastChecked = now
		saveUpdateCache(cache)
	}

	if isNewerVersion(Version, cache.LatestVersion) {
		fmt.Println()
		color.New(color.FgYellow, color.Bold).Printf("💡 A new version of QuotaSense CLI is available: %s (current: %s)\n", cache.LatestVersion, Version)

		// Check if Stdin is a terminal/TTY to avoid hanging in scripting/CI environments
		if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
			color.New(color.FgCyan).Print("👉 Do you want to update now? (y/n): ")

			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				fmt.Println("Fetching update details...")
				latest, err := getLatestRelease()
				if err != nil {
					errorColor.Printf("Error checking for updates: %v\n", err)
					return
				}

				err = doUpdate(latest)
				if err != nil {
					errorColor.Printf("Update failed: %v\n", err)
					return
				}
				successColor.Printf("Successfully updated to %s!\n", latest.TagName)
			} else {
				fmt.Println("Update skipped.")
			}
		} else {
			color.New(color.FgCyan).Printf("👉 Run '%s' to update.\n", "qs update")
		}
		fmt.Println()
	}
}
