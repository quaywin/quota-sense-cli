package models

type AuthFile struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Provider    string `json:"provider"`
	Disabled    bool   `json:"disabled"`
	Unavailable bool   `json:"unavailable"`
	AuthIndex   string `json:"auth_index"`
	ProjectID   string `json:"project_id"`
	Account     string `json:"account"`
}

type AuthFilesResponse struct {
	Files []AuthFile `json:"files"`
}

type ProxyRequest struct {
	AuthIndex string            `json:"authIndex"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Header    map[string]string `json:"header"`
	Data      string            `json:"data"`
}

type ProxyResponse struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

type ModelLimit struct {
	Remaining         string  `json:"remaining"`
	RemainingFraction float64 `json:"remainingFraction"`
	ResetTime         string  `json:"resetTime"`
}

// Google response structures
type GoogleQuotaInfo struct {
	RemainingFraction float64 `json:"remainingFraction"`
	ResetTime         string  `json:"resetTime"`
}

type GoogleModel struct {
	DisplayName string           `json:"displayName"`
	QuotaInfo   *GoogleQuotaInfo `json:"quotaInfo"`
}

type FetchAvailableModelsResponse struct {
	Models map[string]GoogleModel `json:"models"`
}

// Gemini response structures
type GeminiBucket struct {
	ModelID           string  `json:"modelId"`
	RemainingFraction float64 `json:"remainingFraction"`
	ResetTime         string  `json:"resetTime"`
}

type GeminiQuotaResponse struct {
	Buckets []GeminiBucket `json:"buckets"`
}
