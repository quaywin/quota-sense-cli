package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/quaywin/quota-sense-cli/internal/config"
	"github.com/quaywin/quota-sense-cli/internal/models"
)

const (
	googleCloudCodeURL   = "https://daily-cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels"
	geminiQuotaURL       = "https://cloudcode-pa.googleapis.com/v1internal:retrieveUserQuota"
	chatgptUsageURL      = "https://chatgpt.com/backend-api/wham/usage"
	antigravityUserAgent = "antigravity/1.11.5 darwin/amd64"
	codexUserAgent       = "codex_cli_rs/0.76.0 (Debian 13.0.0; x86_64) WindowsTerminal"
)

type Client struct {
	cfg *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) CheckConnection() error {
	_, err := c.FetchUsage()
	if err != nil {
		return fmt.Errorf("failed to connect to server or invalid token: %v", err)
	}
	return nil
}

func (c *Client) FetchUsage() ([]models.AuthFile, error) {
	url := fmt.Sprintf("%s/v0/management/auth-files", c.cfg.ServerURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.cfg.ManagementToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch auth files: %d", resp.StatusCode)
	}

	var authFilesResponse models.AuthFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&authFilesResponse); err != nil {
		return nil, err
	}

	return authFilesResponse.Files, nil
}

func (c *Client) FetchQuota(file models.AuthFile) (map[string]models.ModelLimit, error) {
	if file.Provider == "codex" {
		return c.fetchCodexQuota(file)
	}
	return c.fetchGoogleQuota(file)
}

func (c *Client) fetchCodexQuota(file models.AuthFile) (map[string]models.ModelLimit, error) {
	resp, err := c.GetCodexProvider(file)
	if err != nil {
		return nil, err
	}

	limits := make(map[string]models.ModelLimit)
	modelName := resp.PlanType
	if modelName == "" {
		modelName = "codex"
	}

	remaining := 100.0 - resp.RateLimit.PrimaryWindow.UsedPercent
	if remaining < 0 {
		remaining = 0
	}

	resetTime := time.Unix(resp.RateLimit.PrimaryWindow.ResetAt, 0).Format(time.RFC3339)

	limits[modelName] = models.ModelLimit{
		Remaining:         fmt.Sprintf("%d%%", int(remaining)),
		RemainingFraction: remaining / 100.0,
		ResetTime:         resetTime,
	}
	return limits, nil
}

func (c *Client) fetchGoogleQuota(file models.AuthFile) (map[string]models.ModelLimit, error) {
	proxyURL := fmt.Sprintf("%s/v0/management/api-call", c.cfg.ServerURL)

	targetURL := googleCloudCodeURL
	headers := map[string]string{
		"Authorization": "Bearer $TOKEN$",
		"Content-Type":  "application/json",
		"User-Agent":    antigravityUserAgent,
	}

	data := "{}"
	if file.Provider == "gemini-cli" {
		targetURL = geminiQuotaURL
		delete(headers, "User-Agent")

		projectID := file.ProjectID
		if projectID == "" && file.Account != "" {
			if start := strings.Index(file.Account, "("); start != -1 {
				if end := strings.Index(file.Account, ")"); end != -1 && end > start {
					projectID = file.Account[start+1 : end]
				}
			}
		}
		data = fmt.Sprintf(`{"project":"%s"}`, projectID)
	}

	proxyReqBody := models.ProxyRequest{
		AuthIndex: file.AuthIndex,
		Method:    "POST",
		URL:       targetURL,
		Header:    headers,
		Data:      data,
	}

	jsonData, _ := json.Marshal(proxyReqBody)
	req, _ := http.NewRequest("POST", proxyURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+c.cfg.ManagementToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var proxyResp models.ProxyResponse
	if err := json.NewDecoder(resp.Body).Decode(&proxyResp); err != nil {
		return nil, err
	}

	if proxyResp.StatusCode != 200 {
		return nil, fmt.Errorf("proxy returned status %d", proxyResp.StatusCode)
	}

	limits := make(map[string]models.ModelLimit)

	if file.Provider == "gemini-cli" {
		var geminiResp models.GeminiQuotaResponse
		if err := json.Unmarshal([]byte(proxyResp.Body), &geminiResp); err != nil {
			return nil, err
		}
		for _, bucket := range geminiResp.Buckets {
			if bucket.ModelID != "" {
				limits[bucket.ModelID] = models.ModelLimit{
					Remaining:         fmt.Sprintf("%d%%", int(bucket.RemainingFraction*100)),
					RemainingFraction: bucket.RemainingFraction,
					ResetTime:         bucket.ResetTime,
				}
			}
		}
	} else {
		var googleResp models.FetchAvailableModelsResponse
		if err := json.Unmarshal([]byte(proxyResp.Body), &googleResp); err != nil {
			return nil, err
		}
		for key, model := range googleResp.Models {
			if model.QuotaInfo != nil {
				limits[key] = models.ModelLimit{
					Remaining:         fmt.Sprintf("%d%%", int(model.QuotaInfo.RemainingFraction*100)),
					RemainingFraction: model.QuotaInfo.RemainingFraction,
					ResetTime:         model.QuotaInfo.ResetTime,
				}
			}
		}
	}

	return limits, nil
}

func (c *Client) GetCodexProvider(file models.AuthFile) (*models.CodexUsageResponse, error) {
	proxyURL := fmt.Sprintf("%s/v0/management/api-call", c.cfg.ServerURL)

	headers := map[string]string{
		"Authorization":      "Bearer $TOKEN$",
		"Content-Type":       "application/json",
		"User-Agent":         codexUserAgent,
		"Chatgpt-Account-Id": file.IDToken.ChatgptAccountID,
	}

	proxyReqBody := models.ProxyRequest{
		AuthIndex: file.AuthIndex,
		Method:    "GET",
		URL:       chatgptUsageURL,
		Header:    headers,
	}

	jsonData, err := json.Marshal(proxyReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proxy request: %v", err)
	}

	req, _ := http.NewRequest("POST", proxyURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+c.cfg.ManagementToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("management server returned status %d", resp.StatusCode)
	}

	var proxyResp models.ProxyResponse
	if err := json.NewDecoder(resp.Body).Decode(&proxyResp); err != nil {
		return nil, err
	}

	if proxyResp.StatusCode != 200 {
		return nil, fmt.Errorf("target API returned status %d", proxyResp.StatusCode)
	}

	var usageResp models.CodexUsageResponse
	if err := json.Unmarshal([]byte(proxyResp.Body), &usageResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal usage response: %v", err)
	}

	return &usageResp, nil
}
