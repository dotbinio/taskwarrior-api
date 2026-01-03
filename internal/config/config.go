package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Taskwarrior TaskwarriorConfig `yaml:"taskwarrior"`
	Auth        AuthConfig        `yaml:"auth"`
	Logging     LoggingConfig     `yaml:"logging"`
	CORS        CORSConfig        `yaml:"cors"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	EnableUI bool   `yaml:"enable_ui"`
}

// TaskwarriorConfig holds Taskwarrior-specific configuration
type TaskwarriorConfig struct {
	DataLocation string `yaml:"data_location"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Tokens []string `yaml:"tokens"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		// Set defaults
		Server: ServerConfig{
			Host:     "0.0.0.0",
			Port:     8080,
			EnableUI: true,
		},
		Taskwarrior: TaskwarriorConfig{
			DataLocation: "~/.task",
		},
		Auth: AuthConfig{
			Tokens: []string{},
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		CORS: CORSConfig{
			Enabled:        true,
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	// Load from environment variables
	loadFromEnv(config)

	// Validate configuration
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Server configuration
	if host := os.Getenv("TW_API_HOST"); host != "" {
		config.Server.Host = host
	}
	if portStr := os.Getenv("TW_API_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Server.Port = port
		}
	}
	if enableUIStr := os.Getenv("TW_API_ENABLE_UI"); enableUIStr != "" {
		config.Server.EnableUI = enableUIStr == "true" || enableUIStr == "1"
	}

	// Taskwarrior configuration
	if dataLocation := os.Getenv("TW_DATA_LOCATION"); dataLocation != "" {
		config.Taskwarrior.DataLocation = dataLocation
	}

	// Auth configuration
	if tokensStr := os.Getenv("TW_API_TOKENS"); tokensStr != "" {
		tokens := strings.Split(tokensStr, ",")
		for i, token := range tokens {
			tokens[i] = strings.TrimSpace(token)
		}
		config.Auth.Tokens = tokens
	}

	// Logging configuration
	if level := os.Getenv("TW_API_LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}

	// CORS configuration
	if enabledStr := os.Getenv("TW_API_CORS_ENABLED"); enabledStr != "" {
		config.CORS.Enabled = enabledStr == "true" || enabledStr == "1"
	}
	if originsStr := os.Getenv("TW_API_CORS_ORIGINS"); originsStr != "" {
		origins := strings.Split(originsStr, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		config.CORS.AllowedOrigins = origins
	}
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", config.Server.Port)
	}

	if config.Taskwarrior.DataLocation == "" {
		return fmt.Errorf("taskwarrior data location is required")
	}

	if len(config.Auth.Tokens) == 0 {
		return fmt.Errorf("at least one auth token is required")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", config.Logging.Level)
	}

	return nil
}

// GetAddress returns the full server address
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
