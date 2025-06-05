package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dante-gpu/dante-backend/provider-daemon/internal/billing"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/executor"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/gpu"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/nats"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration for the provider daemon.
type Config struct {
	InstanceID string `yaml:"instance_id"`
	LogLevel   string `yaml:"log_level"`
	// General request timeout, e.g., for HTTP calls to other services
	RequestTimeout time.Duration `yaml:"request_timeout"`

	// NATS Configuration
	NatsConfig         nats.Config     `yaml:"nats"`
	ExecutorConfig     executor.Config `yaml:"executor"`
	GPUDetectorConfig  gpu.Config      `yaml:"gpu_detector"`

	// Provider Registry Configuration (for heartbeats)
	ProviderRegistryServiceName string        `yaml:"provider_registry_service_name"`
	ProviderRegistryURL         string        `yaml:"provider_registry_url,omitempty"`
	ProviderHeartbeatInterval   time.Duration `yaml:"provider_heartbeat_interval"`

	// Task Execution Configuration (Placeholders - to be expanded)
	WorkspaceDir   string   `yaml:"workspace_dir,omitempty"`
	DockerEndpoint string   `yaml:"docker_endpoint,omitempty"`
	ManagedGPUIDs  []string `yaml:"managed_gpu_ids,omitempty"`

	MaxConcurrentJobs uint32 `yaml:"max_concurrent_jobs"`
	PreferredCurrency string `yaml:"preferred_currency"`

	// GpuRentalConfigs holds specific rental settings for each GPU managed by this daemon.
	// These can be initially populated from config.yaml and will be updatable via GUI commands.
	GpuRentalConfigs []GpuRentalConfigEntry `yaml:"gpu_rental_configs,omitempty"`

	Logger              *zap.Logger    `yaml:"-"`
	BillingClientConfig billing.Config  `yaml:"billing_client"`

	shutdownTimeout time.Duration
}

// GpuRentalConfigEntry defines the structure for an individual GPU's rental settings.
type GpuRentalConfigEntry struct {
	GpuID                 string  `yaml:"gpu_id"` // Matches the ID from gpu.DetectedGPU
	IsAvailableForRent    bool    `yaml:"is_available_for_rent"`
	CurrentHourlyRateDGPU float32 `yaml:"current_hourly_rate_dgpu"` // Rate in DGPU per hour
}

// LoadConfig reads configuration from the given YAML file path.
// It creates a default config file if it doesn't exist.
func LoadConfig(path string, logger *zap.Logger) (*Config, error) {
	hostname, _ := os.Hostname()
	defaultInstanceID := "provider-" + hostname
	if defaultInstanceID == "provider-" { // Fallback if hostname fails
		defaultInstanceID = "provider-daemon-unknown"
	}

	defaultConfig := &Config{
		InstanceID:                         defaultInstanceID,
		LogLevel:                           "info",
		RequestTimeout:                     30 * time.Second,
		NatsConfig: nats.Config{
			URL:             "nats://localhost:4222",
			ConnectTimeout:  5 * time.Second,
			ReconnectWait:   5 * time.Second,
			MaxReconnects:   -1, // Infinite
			SubjectPrefix:   "dante.tasks",
			TaskDispatchSubjectPattern: "dante.tasks.dispatch.>", // Listen for tasks dispatched to any provider or this specific one
			StatusUpdateSubject:        "dante.provider.status",
			StreamNamePrefix:           "DANTE_TASKS_", // Prefix for stream names
			ConsumerNamePrefix:         "provider_daemon_", // Prefix for consumer names
			PullMaxWaiting:             5, // Max number of outstanding pull requests for a pull consumer
			AckWait:                    30 * time.Second, // How long to wait for an ack for a message
			MaxDeliver:                 5, // Max number of times to redeliver a message
		},
		ExecutorConfig: executor.Config{
			Type:           "docker", // "docker" or "script"
			DockerEndpoint: "unix:///var/run/docker.sock",
		},
		GPUDetectorConfig: gpu.Config{
			DetectionInterval: 1 * time.Minute,
			NvidiaSmiPath:     "nvidia-smi", // Default path for nvidia-smi
		},
		ProviderRegistryServiceName:        "provider-registry",
		ProviderHeartbeatInterval:          30 * time.Second,
		WorkspaceDir:                       filepath.Join(os.TempDir(), "dante_tasks"),
		MaxConcurrentJobs:                  1,
		PreferredCurrency:                  "DGPU",
		GpuRentalConfigs:                   make([]GpuRentalConfigEntry, 0),
		BillingClientConfig: billing.Config{
			BaseURL: "http://localhost:8081/api/v1/billing",
		},
		shutdownTimeout: 10 * time.Second,
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

	cfg.Logger = logger

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
	if cfg.NatsConfig.URL == "" {
		cfg.NatsConfig.URL = defaults.NatsConfig.URL
	}
	if cfg.ExecutorConfig.Type == "" {
		cfg.ExecutorConfig.Type = defaults.ExecutorConfig.Type
	}
	if cfg.GPUDetectorConfig.DetectionInterval == 0 {
		cfg.GPUDetectorConfig.DetectionInterval = defaults.GPUDetectorConfig.DetectionInterval
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
	if cfg.MaxConcurrentJobs == 0 {
		cfg.MaxConcurrentJobs = defaults.MaxConcurrentJobs
	}
	if cfg.PreferredCurrency == "" {
		cfg.PreferredCurrency = defaults.PreferredCurrency
	}
	if cfg.GpuRentalConfigs == nil {
		cfg.GpuRentalConfigs = defaults.GpuRentalConfigs
	}
	if cfg.BillingClientConfig.BaseURL == "" {
		cfg.BillingClientConfig.BaseURL = defaults.BillingClientConfig.BaseURL
	}
	if cfg.shutdownTimeout == 0 {
		cfg.shutdownTimeout = defaults.shutdownTimeout
	}
}
