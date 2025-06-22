package models

import "time"

type WebhookRequest struct {
	Prompt string `json:"prompt"`
}

type WebhookResponse struct {
	Success       bool   `json:"success"`
	Timestamp     string `json:"timestamp"`
	Prompt        string `json:"prompt"`
	Response      string `json:"response"`
	ExecutionTime string `json:"execution_time"`
	Error         string `json:"error,omitempty"`
}

func NewWebhookResponse(prompt string, success bool) *WebhookResponse {
	return &WebhookResponse{
		Success:   success,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Prompt:    prompt,
	}
}