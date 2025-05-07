package models

import (
	"time"

	"github.com/google/uuid"
)

// ProviderStatus represents the possible states of a GPU provider.
type ProviderStatus string

const (
	StatusIdle        ProviderStatus = "idle"
	StatusBusy        ProviderStatus = "busy"
	StatusOffline     ProviderStatus = "offline"
	StatusMaintenance ProviderStatus = "maintenance"
	StatusError       ProviderStatus = "error"
)

// GPUDetail holds specific information about a GPU.
type GPUDetail struct {
	ModelName     string `json:"model_name" yaml:"model_name"`
	VRAM          uint64 `json:"vram_mb" yaml:"vram_mb"` // VRAM in Megabytes
	DriverVersion string `json:"driver_version" yaml:"driver_version"`
	// We could add more details like CUDA cores, clock speeds, etc.
}

// Provider represents a registered GPU provider in the system.
// This struct will be used for API requests/responses and internal representation.
// For database storage, it would map to a table.
type Provider struct {
	ID           uuid.UUID      `json:"id" yaml:"id"`
	OwnerID      string         `json:"owner_id" yaml:"owner_id"` // ID of the user who owns/registered this provider
	Name         string         `json:"name" yaml:"name"`         // A user-friendly name for the provider rig
	Hostname     string         `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	IPAddress    string         `json:"ip_address,omitempty" yaml:"ip_address,omitempty"`
	Status       ProviderStatus `json:"status" yaml:"status"`
	GPUs         []GPUDetail    `json:"gpus" yaml:"gpus"`                             // A provider can have multiple GPUs
	Location     string         `json:"location,omitempty" yaml:"location,omitempty"` // e.g., "us-east-1a", "home-office-london"
	RegisteredAt time.Time      `json:"registered_at" yaml:"registered_at"`
	LastSeenAt   time.Time      `json:"last_seen_at" yaml:"last_seen_at"`
	// Additional metadata can be stored as a map or a JSONB field in a DB
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// NewProvider creates a new Provider instance with a generated ID and timestamps.
func NewProvider(ownerID, name, hostname, ipAddress, location string, gpus []GPUDetail, metadata map[string]interface{}) *Provider {
	now := time.Now().UTC()
	return &Provider{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         name,
		Hostname:     hostname,
		IPAddress:    ipAddress,
		Status:       StatusIdle, // Default to idle on registration
		GPUs:         gpus,
		Location:     location,
		RegisteredAt: now,
		LastSeenAt:   now,
		Metadata:     metadata,
	}
}

// UpdateStatus updates the provider's status and last seen time.
func (p *Provider) UpdateStatus(newStatus ProviderStatus) {
	p.Status = newStatus
	p.LastSeenAt = time.Now().UTC()
}

// Heartbeat updates the provider's last seen time.
func (p *Provider) Heartbeat() {
	p.LastSeenAt = time.Now().UTC()
	// Optionally, if status was offline/error, a heartbeat might set it to idle
	if p.Status == StatusOffline || p.Status == StatusError {
		p.Status = StatusIdle
	}
}
