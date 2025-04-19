// Package config provides configuration loading and validation functionality
// for the Dead Man's Switch application. It handles loading configuration
// from environment variables and command line flags.
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

	// Telegram bot username
	TelegramBotUsername string

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

	// Load and validate required settings
	if err := loadRequiredSettings(config); err != nil {
		return nil, err
	}

	// Load optional SMTP settings
	if err := loadSMTPSettings(config); err != nil {
		return nil, err
	}

	// Load ping settings
	if err := loadPingSettings(config); err != nil {
		return nil, err
	}

	// Load database settings
	loadDatabaseSettings(config)

	// Load debug and logging settings
	loadDebugSettings(config)

	return config, nil
}

// loadRequiredSettings loads and validates required configuration settings
func loadRequiredSettings(config *Config) error {
	config.BaseDomain = os.Getenv("BASE_DOMAIN")
	config.TelegramBotToken = os.Getenv("TG_BOT_TOKEN")
	config.AdminEmail = os.Getenv("ADMIN_EMAIL")

	if config.BaseDomain == "" {
		return fmt.Errorf("BASE_DOMAIN environment variable is required")
	}

	if config.TelegramBotToken == "" {
		return fmt.Errorf("TG_BOT_TOKEN environment variable is required")
	}

	if config.AdminEmail == "" {
		return fmt.Errorf("ADMIN_EMAIL environment variable is required")
	}

	return nil
}

// loadSMTPSettings loads SMTP configuration settings
func loadSMTPSettings(config *Config) error {
	config.SMTPHost = os.Getenv("SMTP_HOST")

	// Parse SMTP port
	smtpPortStr := os.Getenv("SMTP_PORT")
	if smtpPortStr == "" {
		config.SMTPPort = 587 // Default SMTP port
	} else {
		port, err := strconv.Atoi(smtpPortStr)
		if err != nil {
			return fmt.Errorf("invalid SMTP_PORT: %w", err)
		}
		config.SMTPPort = port
	}

	// Load credentials
	config.SMTPUsername = os.Getenv("SMTP_USERNAME")
	config.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	config.SMTPFrom = os.Getenv("SMTP_FROM")

	// Use username as sender if not specified
	if config.SMTPFrom == "" && config.SMTPUsername != "" {
		config.SMTPFrom = config.SMTPUsername
	}

	return nil
}

// loadPingSettings loads ping frequency and deadline configuration
func loadPingSettings(config *Config) error {
	// Parse ping frequency
	if err := parsePingFrequency(config); err != nil {
		return err
	}

	// Parse ping deadline
	if err := parsePingDeadline(config); err != nil {
		return err
	}

	return nil
}

// parsePingFrequency parses and validates the ping frequency setting
func parsePingFrequency(config *Config) error {
	pingFrequencyStr := os.Getenv("PING_FREQUENCY")
	if pingFrequencyStr == "" {
		config.PingFrequency = 3 * 24 * time.Hour // 3 days default
		return nil
	}

	days, err := strconv.Atoi(pingFrequencyStr)
	if err != nil {
		return fmt.Errorf("invalid PING_FREQUENCY: %w", err)
	}
	if days < 1 || days > 7 {
		return fmt.Errorf("PING_FREQUENCY must be between 1 and 7 days")
	}

	config.PingFrequency = time.Duration(days) * 24 * time.Hour
	return nil
}

// parsePingDeadline parses and validates the ping deadline setting
func parsePingDeadline(config *Config) error {
	pingDeadlineStr := os.Getenv("PING_DEADLINE")
	if pingDeadlineStr == "" {
		config.PingDeadline = 14 * 24 * time.Hour // 14 days default
		return nil
	}

	days, err := strconv.Atoi(pingDeadlineStr)
	if err != nil {
		return fmt.Errorf("invalid PING_DEADLINE: %w", err)
	}
	if days < 7 || days > 30 {
		return fmt.Errorf("PING_DEADLINE must be between 7 and 30 days")
	}

	config.PingDeadline = time.Duration(days) * 24 * time.Hour
	return nil
}

// loadDatabaseSettings loads database configuration settings
func loadDatabaseSettings(config *Config) {
	config.DBPath = os.Getenv("DB_PATH")
	if config.DBPath == "" {
		config.DBPath = "/app/data/db.sqlite"
	}
}

// loadDebugSettings loads debug and logging configuration settings
func loadDebugSettings(config *Config) {
	// Debug mode
	debugStr := os.Getenv("DEBUG")
	config.Debug = debugStr == "true" || debugStr == "1"

	// Log level
	config.LogLevel = os.Getenv("LOG_LEVEL")
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
}

// Validate ensures the configuration is valid
func (c *Config) Validate() error {
	// Check if ping deadline is greater than frequency
	if c.PingDeadline <= c.PingFrequency {
		return fmt.Errorf("PING_DEADLINE must be greater than PING_FREQUENCY")
	}

	return nil
}
