package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Client represents a client for the billing service
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// Config represents billing client configuration
type Config struct {
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// NewClient creates a new billing service client
func NewClient(config *Config, logger *zap.Logger) *Client {
	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// UsageUpdateRequest represents a usage update request
type UsageUpdateRequest struct {
	SessionID        uuid.UUID `json:"session_id"`
	GPUUtilization   uint8     `json:"gpu_utilization_percent"`
	VRAMUtilization  uint8     `json:"vram_utilization_percent"`
	PowerDraw        uint32    `json:"power_draw_w"`
	Temperature      uint8     `json:"temperature_c"`
	Timestamp        time.Time `json:"timestamp"`
}

// SessionResponse represents a session response from billing service
type SessionResponse struct {
	Session struct {
		ID                uuid.UUID       `json:"id"`
		UserID            string          `json:"user_id"`
		ProviderID        uuid.UUID       `json:"provider_id"`
		JobID             *string         `json:"job_id,omitempty"`
		Status            string          `json:"status"`
		GPUModel          string          `json:"gpu_model"`
		AllocatedVRAM     uint64          `json:"allocated_vram_mb"`
		TotalVRAM         uint64          `json:"total_vram_mb"`
		VRAMPercentage    decimal.Decimal `json:"vram_percentage"`
		HourlyRate        decimal.Decimal `json:"hourly_rate"`
		VRAMRate          decimal.Decimal `json:"vram_rate"`
		PowerRate         decimal.Decimal `json:"power_rate"`
		PlatformFeeRate   decimal.Decimal `json:"platform_fee_rate"`
		EstimatedPowerW   uint32          `json:"estimated_power_w"`
		ActualPowerW      *uint32         `json:"actual_power_w,omitempty"`
		StartedAt         time.Time       `json:"started_at"`
		EndedAt           *time.Time      `json:"ended_at,omitempty"`
		LastBilledAt      time.Time       `json:"last_billed_at"`
		TotalCost         decimal.Decimal `json:"total_cost"`
		PlatformFee       decimal.Decimal `json:"platform_fee"`
		ProviderEarnings  decimal.Decimal `json:"provider_earnings"`
		CreatedAt         time.Time       `json:"created_at"`
		UpdatedAt         time.Time       `json:"updated_at"`
	} `json:"session"`
	CurrentCost         decimal.Decimal `json:"current_cost"`
	EstimatedHourlyCost decimal.Decimal `json:"estimated_hourly_cost"`
	RemainingBalance    decimal.Decimal `json:"remaining_balance"`
	EstimatedRuntime    decimal.Decimal `json:"estimated_runtime_hours"`
}

// SendUsageUpdate sends real-time usage data to the billing service
func (c *Client) SendUsageUpdate(ctx context.Context, req *UsageUpdateRequest) error {
	c.logger.Debug("Sending usage update",
		zap.String("session_id", req.SessionID.String()),
		zap.Uint8("gpu_utilization", req.GPUUtilization),
		zap.Uint32("power_draw", req.PowerDraw),
	)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal usage update: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/billing/usage-update", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send usage update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("billing service returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Usage update sent successfully")
	return nil
}

// GetCurrentUsage gets current usage information for a session
func (c *Client) GetCurrentUsage(ctx context.Context, sessionID uuid.UUID) (*SessionResponse, error) {
	url := fmt.Sprintf("%s/api/v1/billing/current-usage/%s", c.baseURL, sessionID.String())
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("billing service returned status %d", resp.StatusCode)
	}

	var sessionResp SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sessionResp, nil
}

// Monitor starts monitoring a session and sends periodic usage updates
func (c *Client) Monitor(ctx context.Context, sessionID uuid.UUID, gpuID string, interval time.Duration) error {
	c.logger.Info("Starting billing monitor",
		zap.String("session_id", sessionID.String()),
		zap.String("gpu_id", gpuID),
		zap.Duration("interval", interval),
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Billing monitor stopped", zap.String("session_id", sessionID.String()))
			return ctx.Err()
		case <-ticker.C:
			// Get current GPU metrics
			metrics, err := c.getGPUMetrics(gpuID)
			if err != nil {
				c.logger.Error("Failed to get GPU metrics", zap.Error(err))
				continue
			}

			// Send usage update
			req := &UsageUpdateRequest{
				SessionID:       sessionID,
				GPUUtilization:  metrics.Utilization,
				VRAMUtilization: metrics.VRAMUtilization,
				PowerDraw:       metrics.PowerDraw,
				Temperature:     metrics.Temperature,
				Timestamp:       time.Now().UTC(),
			}

			if err := c.SendUsageUpdate(ctx, req); err != nil {
				c.logger.Error("Failed to send usage update", zap.Error(err))
				// Continue monitoring even if one update fails
			}
		}
	}
}

// GPUMetrics represents GPU metrics for billing
type GPUMetrics struct {
	Utilization     uint8  `json:"utilization_percent"`
	VRAMUtilization uint8  `json:"vram_utilization_percent"`
	PowerDraw       uint32 `json:"power_draw_w"`
	Temperature     uint8  `json:"temperature_c"`
}

// getGPUMetrics gets current GPU metrics
func (c *Client) getGPUMetrics(gpuID string) (*GPUMetrics, error) {
	// This would integrate with the GPU detector
	// For now, return mock data
	return &GPUMetrics{
		Utilization:     75,
		VRAMUtilization: 60,
		PowerDraw:       250,
		Temperature:     65,
	}, nil
}

// StartSession notifies the billing service that a session has started
func (c *Client) StartSession(ctx context.Context, sessionID uuid.UUID, jobID string) error {
	c.logger.Info("Notifying billing service of session start",
		zap.String("session_id", sessionID.String()),
		zap.String("job_id", jobID),
	)

	// The session would already be created by the scheduler
	// This is just to confirm the session is active on the provider side
	return nil
}

// EndSession notifies the billing service that a session has ended
func (c *Client) EndSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	c.logger.Info("Notifying billing service of session end",
		zap.String("session_id", sessionID.String()),
		zap.String("reason", reason),
	)

	endReq := map[string]interface{}{
		"session_id": sessionID,
		"reason":     reason,
	}

	jsonData, err := json.Marshal(endReq)
	if err != nil {
		return fmt.Errorf("failed to marshal end session request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/billing/end-session", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("billing service returned status %d", resp.StatusCode)
	}

	c.logger.Info("Session end notification sent successfully")
	return nil
}

// CheckSessionStatus checks if a session is still active and funded
func (c *Client) CheckSessionStatus(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	usage, err := c.GetCurrentUsage(ctx, sessionID)
	if err != nil {
		return false, err
	}

	// Check if session is still active and has remaining balance
	isActive := usage.Session.Status == "active" && usage.RemainingBalance.GreaterThan(decimal.Zero)
	
	if !isActive {
		c.logger.Warn("Session is no longer active or funded",
			zap.String("session_id", sessionID.String()),
			zap.String("status", usage.Session.Status),
			zap.String("remaining_balance", usage.RemainingBalance.String()),
		)
	}

	return isActive, nil
}
