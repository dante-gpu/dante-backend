package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// I need a struct to hold my configuration values.
type Config struct {
	Port           string        `yaml:"port"`
	ConsulAddress  string        `yaml:"consul_address"`
	NatsAddress    string        `yaml:"nats_address"`
	LogLevel       string        `yaml:"log_level"`
	JwtSecret      string        `yaml:"jwt_secret"`
	JwtExpiration  time.Duration `yaml:"jwt_expiration"`  // I'll store this as duration
	RequestTimeout time.Duration `yaml:"request_timeout"` // Adding the request timeout here
}

// LoadConfig reads configuration from the given YAML file path.
// It creates a default config file if it doesn't exist.
func LoadConfig(path string) (*Config, error) {
	// I should set some sensible defaults first.
	defaultConfig := &Config{
		Port:           ":8080",
		ConsulAddress:  "localhost:8500",
		NatsAddress:    "nats://localhost:4222", // Using nats:// prefix
		LogLevel:       "info",
		JwtSecret:      "default-very-secure-jwt-secret-key-change-in-production",
		JwtExpiration:  60 * time.Minute, // Defaulting to 60 minutes
		RequestTimeout: 60 * time.Second, // Defaulting to 60 seconds
	}

	// I need to check if the config file exists.
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// If it doesn't exist, I should create it with defaults.
		data, marshalErr := yaml.Marshal(defaultConfig)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal default config: %w", marshalErr)
		}
		// I need to ensure the directory exists before writing.
		if mkdirErr := os.MkdirAll(filepath.Dir(path), 0755); mkdirErr != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", mkdirErr)
		}
		if writeErr := os.WriteFile(path, data, 0644); writeErr != nil {
			return nil, fmt.Errorf("failed to write default config file: %w", writeErr)
		}
		// I can return the defaults now since I just wrote them.
		return defaultConfig, nil
	} else if err != nil {
		// I should handle other potential errors during stat.
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	// If the file exists, I should read it.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// I'll create a config variable to unmarshal into.
	var cfg Config
	// I need to unmarshal the YAML data.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	// It's good practice to apply defaults for any missing fields after loading.
	// This ensures all fields have values even if the file is incomplete.
	applyDefaultsIfNotSet(&cfg, defaultConfig)

	// I should return the loaded configuration.
	return &cfg, nil
}

// applyDefaultsIfNotSet applies default values to cfg fields if they are zero-valued.
// This is useful if the config file exists but is missing some keys.
func applyDefaultsIfNotSet(cfg *Config, defaults *Config) {
	if cfg.Port == "" {
		cfg.Port = defaults.Port
	}
	if cfg.ConsulAddress == "" {
		cfg.ConsulAddress = defaults.ConsulAddress
	}
	if cfg.NatsAddress == "" {
		cfg.NatsAddress = defaults.NatsAddress
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaults.LogLevel
	}
	if cfg.JwtSecret == "" {
		cfg.JwtSecret = defaults.JwtSecret
	}
	if cfg.JwtExpiration == 0 {
		cfg.JwtExpiration = defaults.JwtExpiration
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = defaults.RequestTimeout
	}
}

// Helper function to create the config directory if it doesn't exist
// This was added to the LoadConfig function directly for simplicity
/*
func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
*/
