package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// Base domain for the application
	BaseDomain string

	// Telegram bot token
	TelegramBotToken string

	// Email configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Admin email for notifications
	AdminEmail string

	// Ping settings
	PingFrequency time.Duration
	PingDeadline  time.Duration

	// Database settings
	DBPath string

	// Debug mode
	Debug bool

	// Log level
	LogLevel string
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := &Config{}

	// Required settings
	config.BaseDomain = os.Getenv("BASE_DOMAIN")
	config.TelegramBotToken = os.Getenv("TG_BOT_TOKEN")
	config.AdminEmail = os.Getenv("ADMIN_EMAIL")

	if config.BaseDomain == "" {
		return nil, fmt.Errorf("BASE_DOMAIN environment variable is required")
	}

	if config.TelegramBotToken == "" {
		return nil, fmt.Errorf("TG_BOT_TOKEN environment variable is required")
	}

	if config.AdminEmail == "" {
		return nil, fmt.Errorf("ADMIN_EMAIL environment variable is required")
	}

	// SMTP settings
	config.SMTPHost = os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	if smtpPortStr == "" {
		config.SMTPPort = 587 // Default SMTP port
	} else {
		port, err := strconv.Atoi(smtpPortStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
		}
		config.SMTPPort = port
	}
	config.SMTPUsername = os.Getenv("SMTP_USERNAME")
	config.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	config.SMTPFrom = os.Getenv("SMTP_FROM")
	if config.SMTPFrom == "" && config.SMTPUsername != "" {
		config.SMTPFrom = config.SMTPUsername
	}

	// Ping settings
	pingFrequencyStr := os.Getenv("PING_FREQUENCY")
	if pingFrequencyStr == "" {
		config.PingFrequency = 3 * 24 * time.Hour // 3 days default
	} else {
		days, err := strconv.Atoi(pingFrequencyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PING_FREQUENCY: %w", err)
		}
		if days < 1 || days > 7 {
			return nil, fmt.Errorf("PING_FREQUENCY must be between 1 and 7 days")
		}
		config.PingFrequency = time.Duration(days) * 24 * time.Hour
	}

	pingDeadlineStr := os.Getenv("PING_DEADLINE")
	if pingDeadlineStr == "" {
		config.PingDeadline = 14 * 24 * time.Hour // 14 days default
	} else {
		days, err := strconv.Atoi(pingDeadlineStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PING_DEADLINE: %w", err)
		}
		if days < 7 || days > 30 {
			return nil, fmt.Errorf("PING_DEADLINE must be between 7 and 30 days")
		}
		config.PingDeadline = time.Duration(days) * 24 * time.Hour
	}

	// Database settings
	config.DBPath = os.Getenv("DB_PATH")
	if config.DBPath == "" {
		config.DBPath = "/app/data/db.sqlite"
	}

	// Debug mode
	debugStr := os.Getenv("DEBUG")
	config.Debug = debugStr == "true" || debugStr == "1"

	// Log level
	config.LogLevel = os.Getenv("LOG_LEVEL")
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	return config, nil
}

// Validate ensures the configuration is valid
func (c *Config) Validate() error {
	// Check if ping deadline is greater than frequency
	if c.PingDeadline <= c.PingFrequency {
		return fmt.Errorf("PING_DEADLINE must be greater than PING_FREQUENCY")
	}

	return nil
}
