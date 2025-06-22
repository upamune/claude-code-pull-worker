package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordWebhookURL string
	Port              string
	APIKey            string
	ClaudeTimeout     time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// .env file is optional
	}

	timeoutStr := os.Getenv("CLAUDE_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "1h"
	}
	
	claudeTimeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		claudeTimeout = 1 * time.Hour
	}

	return &Config{
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
		Port:              os.Getenv("PORT"),
		APIKey:            os.Getenv("API_KEY"),
		ClaudeTimeout:     claudeTimeout,
	}, nil
}