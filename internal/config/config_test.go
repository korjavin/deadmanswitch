package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
	// Save original environment variables to restore later
	originalEnv := make(map[string]string)
	envVars := []string{
		"BASE_DOMAIN", "TG_BOT_TOKEN", "ADMIN_EMAIL",
		"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM",
		"PING_FREQUENCY", "PING_DEADLINE", "DB_PATH", "DEBUG", "LOG_LEVEL",
	}

	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
	}

	// Restore environment variables after test
	defer func() {
		for env, value := range originalEnv {
			if value == "" {
				os.Unsetenv(env)
			} else {
				os.Setenv(env, value)
			}
		}
	}()

	// Test cases
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(*testing.T, *Config)
	}{
		{
			name: "Valid configuration with required fields only",
			envVars: map[string]string{
				"BASE_DOMAIN":   "example.com",
				"TG_BOT_TOKEN":  "test-token",
				"ADMIN_EMAIL":   "admin@example.com",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.BaseDomain != "example.com" {
					t.Errorf("Expected BaseDomain to be 'example.com', got '%s'", cfg.BaseDomain)
				}
				if cfg.TelegramBotToken != "test-token" {
					t.Errorf("Expected TelegramBotToken to be 'test-token', got '%s'", cfg.TelegramBotToken)
				}
				if cfg.AdminEmail != "admin@example.com" {
					t.Errorf("Expected AdminEmail to be 'admin@example.com', got '%s'", cfg.AdminEmail)
				}
				// Check defaults
				if cfg.SMTPPort != 587 {
					t.Errorf("Expected default SMTPPort to be 587, got %d", cfg.SMTPPort)
				}
				if cfg.PingFrequency != 3*24*time.Hour {
					t.Errorf("Expected default PingFrequency to be 3 days, got %v", cfg.PingFrequency)
				}
				if cfg.PingDeadline != 14*24*time.Hour {
					t.Errorf("Expected default PingDeadline to be 14 days, got %v", cfg.PingDeadline)
				}
				if cfg.DBPath != "/app/data/db.sqlite" {
					t.Errorf("Expected default DBPath to be '/app/data/db.sqlite', got '%s'", cfg.DBPath)
				}
				if cfg.LogLevel != "info" {
					t.Errorf("Expected default LogLevel to be 'info', got '%s'", cfg.LogLevel)
				}
			},
		},
		{
			name: "Missing BASE_DOMAIN",
			envVars: map[string]string{
				"TG_BOT_TOKEN": "test-token",
				"ADMIN_EMAIL":  "admin@example.com",
			},
			expectError: true,
		},
		{
			name: "Missing TG_BOT_TOKEN",
			envVars: map[string]string{
				"BASE_DOMAIN":  "example.com",
				"ADMIN_EMAIL":  "admin@example.com",
			},
			expectError: true,
		},
		{
			name: "Missing ADMIN_EMAIL",
			envVars: map[string]string{
				"BASE_DOMAIN":  "example.com",
				"TG_BOT_TOKEN": "test-token",
			},
			expectError: true,
		},
		{
			name: "Invalid SMTP_PORT",
			envVars: map[string]string{
				"BASE_DOMAIN":  "example.com",
				"TG_BOT_TOKEN": "test-token",
				"ADMIN_EMAIL":  "admin@example.com",
				"SMTP_PORT":    "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid PING_FREQUENCY",
			envVars: map[string]string{
				"BASE_DOMAIN":    "example.com",
				"TG_BOT_TOKEN":   "test-token",
				"ADMIN_EMAIL":    "admin@example.com",
				"PING_FREQUENCY": "invalid",
			},
			expectError: true,
		},
		{
			name: "PING_FREQUENCY out of range (too low)",
			envVars: map[string]string{
				"BASE_DOMAIN":    "example.com",
				"TG_BOT_TOKEN":   "test-token",
				"ADMIN_EMAIL":    "admin@example.com",
				"PING_FREQUENCY": "0",
			},
			expectError: true,
		},
		{
			name: "PING_FREQUENCY out of range (too high)",
			envVars: map[string]string{
				"BASE_DOMAIN":    "example.com",
				"TG_BOT_TOKEN":   "test-token",
				"ADMIN_EMAIL":    "admin@example.com",
				"PING_FREQUENCY": "8",
			},
			expectError: true,
		},
		{
			name: "Invalid PING_DEADLINE",
			envVars: map[string]string{
				"BASE_DOMAIN":   "example.com",
				"TG_BOT_TOKEN":  "test-token",
				"ADMIN_EMAIL":   "admin@example.com",
				"PING_DEADLINE": "invalid",
			},
			expectError: true,
		},
		{
			name: "PING_DEADLINE out of range (too low)",
			envVars: map[string]string{
				"BASE_DOMAIN":   "example.com",
				"TG_BOT_TOKEN":  "test-token",
				"ADMIN_EMAIL":   "admin@example.com",
				"PING_DEADLINE": "6",
			},
			expectError: true,
		},
		{
			name: "PING_DEADLINE out of range (too high)",
			envVars: map[string]string{
				"BASE_DOMAIN":   "example.com",
				"TG_BOT_TOKEN":  "test-token",
				"ADMIN_EMAIL":   "admin@example.com",
				"PING_DEADLINE": "31",
			},
			expectError: true,
		},
		{
			name: "Full valid configuration",
			envVars: map[string]string{
				"BASE_DOMAIN":    "example.com",
				"TG_BOT_TOKEN":   "test-token",
				"ADMIN_EMAIL":    "admin@example.com",
				"SMTP_HOST":      "smtp.example.com",
				"SMTP_PORT":      "465",
				"SMTP_USERNAME":  "user@example.com",
				"SMTP_PASSWORD":  "password",
				"SMTP_FROM":      "noreply@example.com",
				"PING_FREQUENCY": "2",
				"PING_DEADLINE":  "10",
				"DB_PATH":        "/custom/path/db.sqlite",
				"DEBUG":          "true",
				"LOG_LEVEL":      "debug",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.SMTPHost != "smtp.example.com" {
					t.Errorf("Expected SMTPHost to be 'smtp.example.com', got '%s'", cfg.SMTPHost)
				}
				if cfg.SMTPPort != 465 {
					t.Errorf("Expected SMTPPort to be 465, got %d", cfg.SMTPPort)
				}
				if cfg.SMTPUsername != "user@example.com" {
					t.Errorf("Expected SMTPUsername to be 'user@example.com', got '%s'", cfg.SMTPUsername)
				}
				if cfg.SMTPPassword != "password" {
					t.Errorf("Expected SMTPPassword to be 'password', got '%s'", cfg.SMTPPassword)
				}
				if cfg.SMTPFrom != "noreply@example.com" {
					t.Errorf("Expected SMTPFrom to be 'noreply@example.com', got '%s'", cfg.SMTPFrom)
				}
				if cfg.PingFrequency != 2*24*time.Hour {
					t.Errorf("Expected PingFrequency to be 2 days, got %v", cfg.PingFrequency)
				}
				if cfg.PingDeadline != 10*24*time.Hour {
					t.Errorf("Expected PingDeadline to be 10 days, got %v", cfg.PingDeadline)
				}
				if cfg.DBPath != "/custom/path/db.sqlite" {
					t.Errorf("Expected DBPath to be '/custom/path/db.sqlite', got '%s'", cfg.DBPath)
				}
				if !cfg.Debug {
					t.Errorf("Expected Debug to be true")
				}
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
				}
			},
		},
		{
			name: "Default SMTP_FROM to SMTP_USERNAME when not provided",
			envVars: map[string]string{
				"BASE_DOMAIN":   "example.com",
				"TG_BOT_TOKEN":  "test-token",
				"ADMIN_EMAIL":   "admin@example.com",
				"SMTP_USERNAME": "user@example.com",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.SMTPFrom != "user@example.com" {
					t.Errorf("Expected SMTPFrom to default to SMTPUsername 'user@example.com', got '%s'", cfg.SMTPFrom)
				}
			},
		},
		{
			name: "Debug mode with '1'",
			envVars: map[string]string{
				"BASE_DOMAIN":   "example.com",
				"TG_BOT_TOKEN":  "test-token",
				"ADMIN_EMAIL":   "admin@example.com",
				"DEBUG":         "1",
			},
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				if !cfg.Debug {
					t.Errorf("Expected Debug to be true when DEBUG=1")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clear environment variables
			for _, env := range envVars {
				os.Unsetenv(env)
			}

			// Set environment variables for this test case
			for key, value := range tc.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			cfg, err := LoadFromEnv()

			// Check if error was expected
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			// Check if unexpected error
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Validate configuration
			if tc.validate != nil {
				tc.validate(t, cfg)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: Config{
				PingFrequency: 3 * 24 * time.Hour,
				PingDeadline:  14 * 24 * time.Hour,
			},
			expectError: false,
		},
		{
			name: "PingDeadline equal to PingFrequency",
			config: Config{
				PingFrequency: 7 * 24 * time.Hour,
				PingDeadline:  7 * 24 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "PingDeadline less than PingFrequency",
			config: Config{
				PingFrequency: 7 * 24 * time.Hour,
				PingDeadline:  3 * 24 * time.Hour,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
