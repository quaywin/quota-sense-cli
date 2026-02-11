package models

type AuthFile struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Provider  string  `json:"provider"`
	Disabled  bool    `json:"disabled"`
	AuthIndex string  `json:"auth_index"`
	ProjectID string  `json:"project_id"`
	Account   string  `json:"account"`
	IDToken   IDToken `json:"id_token"`
}

type IDToken struct {
	ChatgptAccountID string `json:"chatgpt_account_id"`
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

// Codex response structures
type CodexUsageResponse struct {
	UserID    string    `json:"user_id"`
	AccountID string    `json:"account_id"`
	Email     string    `json:"email"`
	PlanType  string    `json:"plan_type"`
	RateLimit RateLimit `json:"rate_limit"`
	Credits   any       `json:"credits"`
	Promo     any       `json:"promo"`
}

type RateLimit struct {
	Allowed       bool          `json:"allowed"`
	LimitReached  bool          `json:"limit_reached"`
	PrimaryWindow WindowDetails `json:"primary_window"`
}

type WindowDetails struct {
	UsedPercent        float64 `json:"used_percent"`
	LimitWindowSeconds int     `json:"limit_window_seconds"`
	ResetAfterSeconds  int     `json:"reset_after_seconds"`
	ResetAt            int64   `json:"reset_at"`
}
