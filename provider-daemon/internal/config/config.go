package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration for the provider daemon.
type Config struct {
	InstanceID string `yaml:"instance_id"`
	LogLevel   string `yaml:"log_level"`
	// General request timeout, e.g., for HTTP calls to other services
	RequestTimeout time.Duration `yaml:"request_timeout"`

	// NATS Configuration
	NatsAddress                        string        `yaml:"nats_address"`
	NatsTaskSubscriptionSubjectPattern string        `yaml:"nats_task_subscription_subject_pattern"`
	NatsJobStatusUpdateSubjectPrefix   string        `yaml:"nats_job_status_update_subject_prefix"`
	NatsCommandTimeout                 time.Duration `yaml:"nats_command_timeout"`

	// Provider Registry Configuration (for heartbeats)
	ProviderRegistryServiceName string        `yaml:"provider_registry_service_name"`
	ProviderRegistryURL         string        `yaml:"provider_registry_url,omitempty"`
	ProviderHeartbeatInterval   time.Duration `yaml:"provider_heartbeat_interval"`

	// Task Execution Configuration (Placeholders - to be expanded)
	WorkspaceDir   string   `yaml:"workspace_dir,omitempty"`
	DockerEndpoint string   `yaml:"docker_endpoint,omitempty"`
	ManagedGPUIDs  []string `yaml:"managed_gpu_ids,omitempty"`
}

// LoadConfig reads configuration from the given YAML file path.
// It creates a default config file if it doesn't exist.
func LoadConfig(path string) (*Config, error) {
	hostname, _ := os.Hostname()
	defaultInstanceID := "provider-" + hostname
	if defaultInstanceID == "provider-" { // Fallback if hostname fails
		defaultInstanceID = "provider-daemon-unknown"
	}

	defaultConfig := &Config{
		InstanceID:                         defaultInstanceID,
		LogLevel:                           "info",
		RequestTimeout:                     30 * time.Second,
		NatsAddress:                        "nats://localhost:4222",
		NatsTaskSubscriptionSubjectPattern: "tasks.dispatch.%s.*", // %s will be InstanceID
		NatsJobStatusUpdateSubjectPrefix:   "jobs.status",
		NatsCommandTimeout:                 10 * time.Second,
		ProviderRegistryServiceName:        "provider-registry",
		ProviderHeartbeatInterval:          30 * time.Second,
		WorkspaceDir:                       filepath.Join(os.TempDir(), "dante_tasks"),
	}

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
		fmt.Printf("Default configuration file created at %s\n", path)
		return defaultConfig, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	applyDefaultsIfNotSet(&cfg, defaultConfig)

	return &cfg, nil
}

// applyDefaultsIfNotSet applies default values to cfg fields if they are zero-valued
// or match the type's zero value (e.g., empty string, 0 for time.Duration).
func applyDefaultsIfNotSet(cfg *Config, defaults *Config) {
	if cfg.InstanceID == "" {
		cfg.InstanceID = defaults.InstanceID
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaults.LogLevel
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = defaults.RequestTimeout
	}
	if cfg.NatsAddress == "" {
		cfg.NatsAddress = defaults.NatsAddress
	}
	if cfg.NatsTaskSubscriptionSubjectPattern == "" {
		cfg.NatsTaskSubscriptionSubjectPattern = defaults.NatsTaskSubscriptionSubjectPattern
	}
	if cfg.NatsJobStatusUpdateSubjectPrefix == "" {
		cfg.NatsJobStatusUpdateSubjectPrefix = defaults.NatsJobStatusUpdateSubjectPrefix
	}
	if cfg.NatsCommandTimeout == 0 {
		cfg.NatsCommandTimeout = defaults.NatsCommandTimeout
	}
	if cfg.ProviderRegistryServiceName == "" && cfg.ProviderRegistryURL == "" {
		cfg.ProviderRegistryServiceName = defaults.ProviderRegistryServiceName
	}
	if cfg.ProviderHeartbeatInterval == 0 {
		cfg.ProviderHeartbeatInterval = defaults.ProviderHeartbeatInterval
	}
	if cfg.WorkspaceDir == "" {
		cfg.WorkspaceDir = defaults.WorkspaceDir
	}
}
