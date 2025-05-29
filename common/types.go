package common

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GPUDetail represents GPU hardware information
type GPUDetail struct {
	ModelName         string `json:"model_name"`
	VRAM              uint64 `json:"vram_mb"`
	DriverVersion     string `json:"driver_version"`
	Architecture      string `json:"architecture,omitempty"`
	ComputeCapability string `json:"compute_capability,omitempty"`
	CudaCores         uint32 `json:"cuda_cores,omitempty"`
	TensorCores       uint32 `json:"tensor_cores,omitempty"`
	MemoryBandwidth   uint64 `json:"memory_bandwidth_gb_s,omitempty"`
	PowerConsumption  uint32 `json:"power_consumption_w,omitempty"`

	// Current metrics
	UtilizationGPU uint8  `json:"utilization_gpu_percent,omitempty"`
	UtilizationMem uint8  `json:"utilization_memory_percent,omitempty"`
	Temperature    uint8  `json:"temperature_celsius,omitempty"`
	PowerDraw      uint32 `json:"power_draw_w,omitempty"`

	// Status
	IsHealthy   bool      `json:"is_healthy"`
	IsAvailable bool      `json:"is_available"`
	LastCheckAt time.Time `json:"last_check_at,omitempty"`
}

// Provider represents a GPU provider in the system
type Provider struct {
	ID         uuid.UUID        `json:"id"`
	OwnerID    string           `json:"owner_id"`
	Name       string           `json:"name"`
	Hostname   string           `json:"hostname"`
	IPAddress  string           `json:"ip_address"`
	Location   string           `json:"location"`
	Status     string           `json:"status"`
	GPUs       []GPUDetail      `json:"gpus"`
	Metadata   ProviderMetadata `json:"metadata"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	LastSeenAt *time.Time       `json:"last_seen_at,omitempty"`
}

// ProviderMetadata contains additional provider information
type ProviderMetadata struct {
	MaxConcurrentJobs   int             `json:"max_concurrent_jobs"`
	MinPricePerHour     decimal.Decimal `json:"min_price_per_hour"`
	SolanaWallet        string          `json:"solana_wallet"`
	DockerEnabled       bool            `json:"docker_enabled"`
	SupportedFrameworks []string        `json:"supported_frameworks,omitempty"`
	Tags                []string        `json:"tags,omitempty"`
}

// ProviderConfig represents configuration for GPU provider
type ProviderConfig struct {
	// Basic provider info
	ProviderName string `json:"provider_name"`
	OwnerID      string `json:"owner_id"`
	Location     string `json:"location"`

	// Service URLs
	APIGatewayURL       string `json:"api_gateway_url"`
	ProviderRegistryURL string `json:"provider_registry_url"`
	BillingServiceURL   string `json:"billing_service_url"`
	NATSAddress         string `json:"nats_address"`

	// Provider settings
	SolanaWalletAddress string          `json:"solana_wallet_address"`
	MaxConcurrentJobs   int             `json:"max_concurrent_jobs"`
	MinPricePerHour     decimal.Decimal `json:"min_price_per_hour"`
	EnableDocker        bool            `json:"enable_docker"`

	// Intervals and timeouts
	RequestTimeout    time.Duration `json:"request_timeout"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	MetricsInterval   time.Duration `json:"metrics_interval"`

	// Optional workspace settings
	WorkspaceDir string `json:"workspace_dir,omitempty"`
}

// GPURentalConfig holds configuration for the GPU rental client
type GPURentalConfig struct {
	APIGatewayURL         string          `json:"api_gateway_url"`
	ProviderRegistryURL   string          `json:"provider_registry_url"`
	BillingServiceURL     string          `json:"billing_service_url"`
	StorageServiceURL     string          `json:"storage_service_url"`
	Username              string          `json:"username"`
	Password              string          `json:"password"`
	SolanaPrivateKey      string          `json:"solana_private_key"`
	DefaultJobType        string          `json:"default_job_type"`
	DefaultMaxCostDGPU    decimal.Decimal `json:"default_max_cost_dgpu"`
	DefaultMaxDurationHrs int             `json:"default_max_duration_hrs"`
	DefaultGPUType        string          `json:"default_gpu_type"`
	DefaultVRAMGB         int             `json:"default_vram_gb"`
	RequestTimeout        time.Duration   `json:"request_timeout"`
	PollingInterval       time.Duration   `json:"polling_interval"`
	EnableAutoRetry       bool            `json:"enable_auto_retry"`
	MaxRetryAttempts      int             `json:"max_retry_attempts"`
	EnableNotifications   bool            `json:"enable_notifications"`
}
