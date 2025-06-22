package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/upamune/claude-code-pull-worker/internal/models"
)

const (
	ColorSuccess = 0x00ff00 // Green
	ColorError   = 0xff0000 // Red
	MaxFieldLen  = 1024
	MaxDescLen   = 2048
)

type Client struct {
	webhookURL string
}

func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
	}
}

func (c *Client) SendNotification(resp *models.WebhookResponse) error {
	color := ColorSuccess
	if !resp.Success {
		color = ColorError
	}

	embed := Embed{
		Title:     "Claude Code Execution Result",
		Timestamp: resp.Timestamp,
		Color:     color,
		Fields: []Field{
			{
				Name:   "Prompt",
				Value:  truncate(resp.Prompt, MaxFieldLen),
				Inline: false,
			},
			{
				Name:   "Execution Time",
				Value:  resp.ExecutionTime,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  fmt.Sprintf("%v", resp.Success),
				Inline: true,
			},
		},
	}

	if resp.Success {
		embed.Description = truncate(resp.Response, MaxDescLen)
	} else {
		embed.Description = fmt.Sprintf("Error: %s", truncate(resp.Error, MaxDescLen))
	}

	webhook := Webhook{
		Embeds: []Embed{embed},
	}

	payload, err := json.Marshal(webhook)
	if err != nil {
		log.Printf("Error marshaling Discord webhook: %v", err)
		return fmt.Errorf("failed to marshal Discord webhook: %w", err)
	}

	httpResp, err := http.Post(c.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error sending Discord notification: %v", err)
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		log.Printf("Discord webhook returned status: %d", httpResp.StatusCode)
		return fmt.Errorf("Discord webhook returned unexpected status: %d", httpResp.StatusCode)
	}

	return nil
}

func (c *Client) Name() string {
	return "discord"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}