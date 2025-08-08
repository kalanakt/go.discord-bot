package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Discord Bot Configuration
	BotToken      string
	CommandPrefix string
	DevGuildID    string // For development/testing with guild-specific commands

	// Database Configuration
	DatabaseURL string

	// Development Mode
	DevMode bool
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, errors.New("BOT_TOKEN environment variable is required")
	}

	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		return nil, errors.New("DB_URL environment variable is required")
	}

	// Optional configuration with defaults
	commandPrefix := os.Getenv("BOT_PREFIX")
	if commandPrefix == "" {
		commandPrefix = "!"
	}

	devMode := false
	devModeStr := strings.ToLower(os.Getenv("BOT_DEV_MODE"))
	if devModeStr == "true" || devModeStr == "1" || devModeStr == "yes" {
		devMode = true
	}

	// Guild ID is optional and only used in development mode
	devGuildID := os.Getenv("BOT_DEV_GUILD_ID")

	return &Config{
		BotToken:      botToken,
		CommandPrefix: commandPrefix,
		DevGuildID:    devGuildID,
		DatabaseURL:   databaseURL,
		DevMode:       devMode,
	}, nil
}
