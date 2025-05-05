package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
// I need settings for the server port, logging, Consul, DB, and service registration.
type Config struct {
	Port                string        `yaml:"port"`
	LogLevel            string        `yaml:"log_level"`
	ConsulAddress       string        `yaml:"consul_address"`
	DatabaseURL         string        `yaml:"database_url"`      // Planned for DB connection
	ServiceName         string        `yaml:"service_name"`      // Name to register with Consul
	ServiceIDPrefix     string        `yaml:"service_id_prefix"` // Prefix for unique Consul service ID
	ServiceTags         []string      `yaml:"service_tags"`      // Tags for Consul registration
	HealthCheckPath     string        `yaml:"health_check_path"` // Path for Consul health check
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	HealthCheckTimeout  time.Duration `yaml:"health_check_timeout"`
	RequestTimeout      time.Duration `yaml:"request_timeout"`
}

// LoadConfig reads configuration from the given YAML file path.
// It creates a default config file if it doesn't exist.
func LoadConfig(path string) (*Config, error) {
	// I should set some sensible defaults first.
	defaultConfig := &Config{
		Port:                ":8002",
		LogLevel:            "info",
		ConsulAddress:       "localhost:8500",
		DatabaseURL:         "postgresql://user:pass@localhost:5432/dante_registry?sslmode=disable",
		ServiceName:         "provider-registry",
		ServiceIDPrefix:     "provider-reg-", // Keep it short
		ServiceTags:         []string{"dante", "registry"},
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  2 * time.Second,
		RequestTimeout:      30 * time.Second,
	}

	// Check if file exists, create if not
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		data, marshalErr := yaml.Marshal(defaultConfig)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal default config: %w", marshalErr)
		}
		if mkdirErr := os.MkdirAll(filepath.Dir(path), 0755); mkdirErr != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", mkdirErr)
		}
		if writeErr := os.WriteFile(path, data, 0644); writeErr != nil {
			return nil, fmt.Errorf("failed to write default config file: %w", writeErr)
		}
		return defaultConfig, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	// Read existing file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	// Apply defaults for any missing fields
	applyDefaultsIfNotSet(&cfg, defaultConfig)

	return &cfg, nil
}

// applyDefaultsIfNotSet applies default values to cfg fields if they are zero-valued.
func applyDefaultsIfNotSet(cfg *Config, defaults *Config) {
	if cfg.Port == "" {
		cfg.Port = defaults.Port
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaults.LogLevel
	}
	if cfg.ConsulAddress == "" {
		cfg.ConsulAddress = defaults.ConsulAddress
	}
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = defaults.DatabaseURL
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = defaults.ServiceName
	}
	if cfg.ServiceIDPrefix == "" {
		cfg.ServiceIDPrefix = defaults.ServiceIDPrefix
	}
	if len(cfg.ServiceTags) == 0 {
		cfg.ServiceTags = defaults.ServiceTags
	}
	if cfg.HealthCheckPath == "" {
		cfg.HealthCheckPath = defaults.HealthCheckPath
	}
	if cfg.HealthCheckInterval == 0 {
		cfg.HealthCheckInterval = defaults.HealthCheckInterval
	}
	if cfg.HealthCheckTimeout == 0 {
		cfg.HealthCheckTimeout = defaults.HealthCheckTimeout
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = defaults.RequestTimeout
	}
}

// Helper function to generate a unique Service ID for Consul
func GenerateServiceID(prefix string) string {
	// I should append a unique part to the prefix.
	// Using a UUID is a good way to ensure uniqueness.
	return prefix + uuid.New().String()
}
