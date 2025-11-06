package config

import (
	"errors"
	"os"
)

// Config holds the application configuration
type Config struct {
	TelegramToken string
	// Add more configuration fields as needed
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, errors.New("TELEGRAM_TOKEN environment variable is required")
	}
	
	cfg := &Config{
		TelegramToken: token,
	}
	return cfg, nil
}
