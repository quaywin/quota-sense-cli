package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ServerURL       string `json:"server_url"`
	ManagementToken string `json:"management_token"`
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".quota-sense.json")
}

func LoadConfig() (*Config, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.ServerURL == "" || cfg.ManagementToken == "" {
		return nil, fmt.Errorf("invalid config")
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path := GetConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func PromptConfig() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	var cfg Config

	fmt.Println("=== QuotaSense Configuration ===")

	fmt.Print("Enter Remote Server URL (e.g., http://localhost:8080): ")
	url, _ := reader.ReadString('\n')
	cfg.ServerURL = strings.TrimSpace(url)

	fmt.Print("Enter Management Token (Secret Key): ")
	token, _ := reader.ReadString('\n')
	cfg.ManagementToken = strings.TrimSpace(token)

	if cfg.ServerURL == "" || cfg.ManagementToken == "" {
		return nil, fmt.Errorf("server URL and Management Token are required")
	}

	return &cfg, nil
}
