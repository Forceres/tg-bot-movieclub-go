package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	TelegramToken string
	// Add more configuration fields as needed
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
	}
	return cfg, nil
}
