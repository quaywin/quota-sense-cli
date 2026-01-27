package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const repoOwner = "quaywin"
const repoName = "quota-sense-cli"

type releaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update QuotaSense CLI to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")
		latest, err := getLatestRelease()
		if err != nil {
			errorColor.Printf("Error checking for updates: %v\n", err)
			return
		}

		if latest.TagName == Version {
			successColor.Printf("You are already on the latest version (%s)\n", Version)
			return
		}

		fmt.Printf("New version available: %s (current: %s)\n", latest.TagName, Version)
		fmt.Print("Do you want to update? (y/n): ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			fmt.Println("Update cancelled.")
			return
		}

		err = doUpdate(latest)
		if err != nil {
			errorColor.Printf("Update failed: %v\n", err)
			return
		}

		successColor.Printf("Successfully updated to %s!\n", latest.TagName)
	},
}

func getLatestRelease() (*releaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func doUpdate(release *releaseInfo) error {
	// Determine target asset name
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	extension := "tar.gz"
	if targetOS == "windows" {
		extension = "zip"
	}

	// Format expected asset name: qs_v0.1.0_darwin_arm64.tar.gz
	assetPattern := fmt.Sprintf("%s_%s_%s", repoName, targetOS, targetArch)
	var downloadURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetPattern) && strings.HasSuffix(asset.Name, extension) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("could not find a compatible binary for %s/%s in release %s", targetOS, targetArch, release.TagName)
	}

	fmt.Println("Downloading latest version...")
	tmpDir, err := os.MkdirTemp("", "qs-update")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "archive."+extension)
	err = downloadFile(archivePath, downloadURL)
	if err != nil {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// Extract binary
	fmt.Println("Installing update...")
	binaryName := "qs"
	if targetOS == "windows" {
		binaryName += ".exe"
	}

	var extractedBinary string
	if extension == "tar.gz" {
		cmd := exec.Command("tar", "-xzf", archivePath, "-C", tmpDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %v", err)
		}
		extractedBinary = filepath.Join(tmpDir, binaryName)
	} else {
		// Simple unzip logic for windows if needed, or assume 'unzip' command
		cmd := exec.Command("tar", "-xf", archivePath, "-C", tmpDir) // Windows 'tar' handles zip too
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract zip: %v", err)
		}
		extractedBinary = filepath.Join(tmpDir, binaryName)
	}

	if _, err := os.Stat(extractedBinary); os.IsNotExist(err) {
		return fmt.Errorf("extracted binary not found in archive")
	}

	// Replace current binary
	if targetOS == "windows" {
		// Windows specific: can't replace running exe
		oldExe := exePath + ".old"
		_ = os.Remove(oldExe)
		if err := os.Rename(exePath, oldExe); err != nil {
			return err
		}
		if err := moveFile(extractedBinary, exePath); err != nil {
			return err
		}
		fmt.Println("Note: On Windows, you might need to manually delete the .old file.")
	} else {
		// Unix: use Rename to replace the binary
		if err := os.Rename(extractedBinary, exePath); err != nil {
			// If rename fails (e.g. cross-device), try moving via copy
			if err := moveFile(extractedBinary, exePath); err != nil {
				return err
			}
		}
		// Ensure executable permissions
		if err := os.Chmod(exePath, 0755); err != nil {
			return err
		}
	}

	return nil
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func moveFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
