package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// GetDisplayModelName returns a user-friendly name for a given model ID and provider.
// If fullMode is true, it returns the original modelName.
func GetDisplayModelName(modelName, provider string, fullMode bool) string {
	if fullMode {
		return modelName
	}

	if provider == "antigravity" {
		lowerModel := strings.ToLower(modelName)
		if strings.Contains(lowerModel, "claude") {
			return "Claude/GPT"
		}
		if strings.Contains(lowerModel, "pro") {
			return "Gemini 3 Pro"
		}
		if strings.Contains(lowerModel, "flash") {
			return "Gemini 3 Flash"
		}
		return "" // Signal to skip this model
	} else if provider == "gemini-cli" {
		lowerModel := strings.ToLower(modelName)
		if strings.Contains(lowerModel, "pro") {
			return "Gemini Pro"
		}
		if strings.Contains(lowerModel, "flash") {
			return "Gemini Flash"
		}
		return "" // Signal to skip this model
	} else if provider == "codex" {
		return strings.Title(modelName)
	}

	return modelName
}

// GetQuotaColor returns a color based on the remaining quota percentage.
func GetQuotaColor(remainingVal int) *color.Color {
	if remainingVal > 50 {
		return color.New(color.FgGreen)
	} else if remainingVal > 20 {
		return color.New(color.FgYellow)
	}
	return color.New(color.FgRed, color.Bold)
}

// FormatDuration formats a duration into a human-readable string (e.g., "2h 30m").
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// GetResetString returns a formatted string for the reset time.
func GetResetString(resetTimeStr string) string {
	if resetTimeStr == "" {
		return "-"
	}
	resetTime, err := time.Parse(time.RFC3339, resetTimeStr)
	if err != nil {
		return "-"
	}
	duration := time.Until(resetTime)
	if duration > 0 {
		return FormatDuration(duration)
	}
	return "Now"
}
