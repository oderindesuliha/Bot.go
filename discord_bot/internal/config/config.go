package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	DiscordBotToken string
	CoralBackendURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	config := &Config{
		DiscordBotToken: os.Getenv("DISCORD_BOT_TOKEN"),
		CoralBackendURL: os.Getenv("CORAL_BACKEND_URL"),
	}

	// Validate required configuration
	if config.DiscordBotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN is required")
	}

	return config
}
