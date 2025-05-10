package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dante-gpu/dante-backend/provider-registry-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgresProviderStore is a PostgreSQL implementation of the ProviderStore interface.
type PostgresProviderStore struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgresProviderStore creates a new PostgresProviderStore.
func NewPostgresProviderStore(db *pgxpool.Pool, logger *zap.Logger) *PostgresProviderStore {
	return &PostgresProviderStore{
		db:     db,
		logger: logger,
	}
}

// Initialize sets up the PostgreSQL tables if they don't exist.
func (pps *PostgresProviderStore) Initialize(ctx context.Context) error {
	// Create providers table if it doesn't exist
	sqlProviders := `
	CREATE TABLE IF NOT EXISTS providers (
		id UUID PRIMARY KEY,
		owner_id TEXT NOT NULL,
		name TEXT NOT NULL,
		hostname TEXT,
		ip_address TEXT,
		status TEXT NOT NULL,
		location TEXT,
		registered_at TIMESTAMPTZ NOT NULL,
		last_seen_at TIMESTAMPTZ NOT NULL,
		metadata JSONB,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- Create indexes on frequently queried columns
	CREATE INDEX IF NOT EXISTS idx_providers_owner_id ON providers(owner_id);
	CREATE INDEX IF NOT EXISTS idx_providers_status ON providers(status);
	CREATE INDEX IF NOT EXISTS idx_providers_name ON providers(name);
	CREATE INDEX IF NOT EXISTS idx_providers_last_seen_at ON providers(last_seen_at);
	`

	// Create GPU details table
	sqlGPU := `
	CREATE TABLE IF NOT EXISTS gpu_details (
		id SERIAL PRIMARY KEY,
		provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
		model_name TEXT NOT NULL,
		vram_mb BIGINT NOT NULL,
		driver_version TEXT,
		architecture TEXT,
		compute_capability TEXT,
		cuda_cores INTEGER,
		tensor_cores INTEGER,
		memory_bandwidth_gb_s INTEGER,
		power_consumption_w INTEGER,
		utilization_gpu_percent SMALLINT,
		utilization_memory_percent SMALLINT,
		temperature_c SMALLINT,
		power_draw_w INTEGER,
		is_healthy BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- Create indexes for GPU details table
	CREATE INDEX IF NOT EXISTS idx_gpu_details_provider_id ON gpu_details(provider_id);
	CREATE INDEX IF NOT EXISTS idx_gpu_details_model_name ON gpu_details(model_name);
	CREATE INDEX IF NOT EXISTS idx_gpu_details_vram ON gpu_details(vram_mb);
	CREATE INDEX IF NOT EXISTS idx_gpu_details_architecture ON gpu_details(architecture);
	CREATE INDEX IF NOT EXISTS idx_gpu_details_is_healthy ON gpu_details(is_healthy);
	`

	// Execute the table creation queries
	if _, err := pps.db.Exec(ctx, sqlProviders); err != nil {
		return fmt.Errorf("failed to create providers table: %w", err)
	}

	if _, err := pps.db.Exec(ctx, sqlGPU); err != nil {
		return fmt.Errorf("failed to create gpu_details table: %w", err)
	}

	pps.logger.Info("PostgreSQL tables initialized for provider store")
	return nil
}

// AddProvider adds a new provider to the PostgreSQL database.
func (pps *PostgresProviderStore) AddProvider(ctx context.Context, provider *models.Provider) error {
	// Start transaction
	tx, err := pps.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert provider record
	sqlProvider := `
	INSERT INTO providers (
		id, owner_id, name, hostname, ip_address, status, location, 
		registered_at, last_seen_at, metadata
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// Convert metadata to JSON if it exists
	var metadataJSON []byte
	var err1 error
	if provider.Metadata != nil {
		metadataJSON, err1 = json.Marshal(provider.Metadata)
		if err1 != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to marshal metadata: %w", err1)
		}
	}

	_, err = tx.Exec(
		ctx,
		sqlProvider,
		provider.ID,
		provider.OwnerID,
		provider.Name,
		provider.Hostname,
		provider.IPAddress,
		provider.Status,
		provider.Location,
		provider.RegisteredAt,
		provider.LastSeenAt,
		metadataJSON,
	)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to insert provider: %w", err)
	}

	// Insert each GPU detail
	sqlGPU := `
	INSERT INTO gpu_details (
		provider_id, model_name, vram_mb, driver_version, 
		architecture, compute_capability, cuda_cores, tensor_cores, 
		memory_bandwidth_gb_s, power_consumption_w, 
		utilization_gpu_percent, utilization_memory_percent, 
		temperature_c, power_draw_w, is_healthy
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	for _, gpu := range provider.GPUs {
		_, err = tx.Exec(ctx, sqlGPU,
			provider.ID,
			gpu.ModelName,
			gpu.VRAM,
			gpu.DriverVersion,
			gpu.Architecture,
			gpu.ComputeCapability,
			gpu.CudaCores,
			gpu.TensorCores,
			gpu.MemoryBandwidth,
			gpu.PowerConsumption,
			gpu.UtilizationGPU,
			gpu.UtilizationMem,
			gpu.Temperature,
			gpu.PowerDraw,
			gpu.IsHealthy,
		)
		if err != nil {
			return fmt.Errorf("failed to insert GPU detail: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetProvider retrieves a provider by its ID.
func (pps *PostgresProviderStore) GetProvider(ctx context.Context, id uuid.UUID) (*models.Provider, error) {
	// Query the provider
	sqlProvider := `
	SELECT 
		id, owner_id, name, hostname, ip_address, status, location, 
		registered_at, last_seen_at, metadata
	FROM providers 
	WHERE id = $1
	`

	var provider models.Provider
	var metadataJSON []byte
	var hostname, ipAddress, location sql.NullString // For handling nullable fields

	err := pps.db.QueryRow(ctx, sqlProvider, id).Scan(
		&provider.ID,
		&provider.OwnerID,
		&provider.Name,
		&hostname,
		&ipAddress,
		&provider.Status,
		&location,
		&provider.RegisteredAt,
		&provider.LastSeenAt,
		&metadataJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to query provider: %w", err)
	}

	// Handle null fields
	if hostname.Valid {
		provider.Hostname = hostname.String
	}
	if ipAddress.Valid {
		provider.IPAddress = ipAddress.String
	}
	if location.Valid {
		provider.Location = location.String
	}

	// Unmarshal metadata if not null
	if len(metadataJSON) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		provider.Metadata = metadata
	}

	// Get GPU details for this provider
	sqlGPU := `
	SELECT model_name, vram_mb, driver_version, 
	       architecture, compute_capability, cuda_cores, tensor_cores, 
	       memory_bandwidth_gb_s, power_consumption_w, 
	       utilization_gpu_percent, utilization_memory_percent, 
	       temperature_c, power_draw_w, is_healthy
	FROM gpu_details 
	WHERE provider_id = $1
	`

	rows, err := pps.db.Query(ctx, sqlGPU, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query GPU details: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var gpu models.GPUDetail
		if err := rows.Scan(
			&gpu.ModelName,
			&gpu.VRAM,
			&gpu.DriverVersion,
			&gpu.Architecture,
			&gpu.ComputeCapability,
			&gpu.CudaCores,
			&gpu.TensorCores,
			&gpu.MemoryBandwidth,
			&gpu.PowerConsumption,
			&gpu.UtilizationGPU,
			&gpu.UtilizationMem,
			&gpu.Temperature,
			&gpu.PowerDraw,
			&gpu.IsHealthy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan GPU detail: %w", err)
		}
		provider.GPUs = append(provider.GPUs, gpu)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating GPU details rows: %w", err)
	}

	return &provider, nil
}

// ListProviders returns a list of all providers, with optional filtering.
func (pps *PostgresProviderStore) ListProviders(ctx context.Context) ([]*models.Provider, error) {
	// Query both provider information and GPU details with a single query
	// Use LEFT JOIN to include providers without GPUs and JSON_AGG to aggregate GPU details
	sqlQuery := `
	SELECT p.id, p.owner_id, p.name, p.hostname, p.ip_address, p.status, p.location, 
	       p.registered_at, p.last_seen_at, p.metadata,
		COALESCE(
			JSON_AGG(
				JSON_BUILD_OBJECT(
					'model_name', g.model_name,
					'vram_mb', g.vram_mb,
					'driver_version', g.driver_version,
					'architecture', g.architecture,
					'compute_capability', g.compute_capability,
					'cuda_cores', g.cuda_cores,
					'tensor_cores', g.tensor_cores,
					'memory_bandwidth_gb_s', g.memory_bandwidth_gb_s,
					'power_consumption_w', g.power_consumption_w,
					'utilization_gpu_percent', g.utilization_gpu_percent,
					'utilization_memory_percent', g.utilization_memory_percent,
					'temperature_c', g.temperature_c,
					'power_draw_w', g.power_draw_w,
					'is_healthy', g.is_healthy
				)
			) FILTER (WHERE g.id IS NOT NULL),
			'[]'::JSON
		) AS gpus
	FROM providers p
	LEFT JOIN gpu_details g ON p.id = g.provider_id
	GROUP BY p.id
	ORDER BY p.name
	`

	rows, err := pps.db.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}
	defer rows.Close()

	providers := []*models.Provider{}
	for rows.Next() {
		var (
			provider     = &models.Provider{}
			metadataJSON []byte
			gpusJSON     []byte
			hostname     string
			ipAddress    string
			location     string
		)

		err := rows.Scan(
			&provider.ID,
			&provider.OwnerID,
			&provider.Name,
			&hostname,
			&ipAddress,
			&provider.Status,
			&location,
			&provider.RegisteredAt,
			&provider.LastSeenAt,
			&metadataJSON,
			&gpusJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		// Set the scanned values
		provider.Hostname = hostname
		provider.IPAddress = ipAddress
		provider.Location = location

		// Unmarshal metadata if it exists
		if len(metadataJSON) > 0 && !strings.EqualFold(string(metadataJSON), "null") {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata for provider %s: %w", provider.ID, err)
			}
			provider.Metadata = metadata
		} else {
			provider.Metadata = make(map[string]interface{})
		}

		// Unmarshal GPU details
		if len(gpusJSON) > 0 && !strings.EqualFold(string(gpusJSON), "null") && !strings.EqualFold(string(gpusJSON), "[]") {
			var gpuDetails []models.GPUDetail
			if err := json.Unmarshal(gpusJSON, &gpuDetails); err != nil {
				return nil, fmt.Errorf("failed to unmarshal GPU details for provider %s: %w", provider.ID, err)
			}
			provider.GPUs = gpuDetails
		} else {
			provider.GPUs = []models.GPUDetail{} // Empty slice instead of nil
		}

		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provider rows: %w", err)
	}

	return providers, nil
}

// UpdateProvider updates an existing provider in the database.
func (pps *PostgresProviderStore) UpdateProvider(ctx context.Context, id uuid.UUID, updatedProvider *models.Provider) error {
	// Start a transaction
	tx, err := pps.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	// Update the provider record
	metadataJSON, err := json.Marshal(updatedProvider.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	sqlProvider := `
	UPDATE providers SET 
		owner_id = $1,
		name = $2,
		hostname = $3,
		ip_address = $4,
		status = $5,
		location = $6,
		last_seen_at = $7,
		metadata = $8
	WHERE id = $9
	`

	result, err := tx.Exec(ctx, sqlProvider,
		updatedProvider.OwnerID,
		updatedProvider.Name,
		updatedProvider.Hostname,
		updatedProvider.IPAddress,
		updatedProvider.Status,
		updatedProvider.Location,
		updatedProvider.LastSeenAt,
		metadataJSON,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return models.ErrProviderNotFound
	}

	// Delete existing GPU details for this provider
	sqlDeleteGPUs := `DELETE FROM gpu_details WHERE provider_id = $1`
	_, err = tx.Exec(ctx, sqlDeleteGPUs, id)
	if err != nil {
		return fmt.Errorf("failed to delete existing GPU details: %w", err)
	}

	// Insert updated GPU details
	sqlGPU := `
	INSERT INTO gpu_details (
		provider_id, model_name, vram_mb, driver_version,
		architecture, compute_capability, cuda_cores, tensor_cores,
		memory_bandwidth_gb_s, power_consumption_w,
		utilization_gpu_percent, utilization_memory_percent,
		temperature_c, power_draw_w, is_healthy
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	for _, gpu := range updatedProvider.GPUs {
		_, err = tx.Exec(ctx, sqlGPU,
			id,
			gpu.ModelName,
			gpu.VRAM,
			gpu.DriverVersion,
			gpu.Architecture,
			gpu.ComputeCapability,
			gpu.CudaCores,
			gpu.TensorCores,
			gpu.MemoryBandwidth,
			gpu.PowerConsumption,
			gpu.UtilizationGPU,
			gpu.UtilizationMem,
			gpu.Temperature,
			gpu.PowerDraw,
			gpu.IsHealthy,
		)
		if err != nil {
			return fmt.Errorf("failed to insert updated GPU detail: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteProvider removes a provider from the database.
func (pps *PostgresProviderStore) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	// The GPU details will be automatically deleted due to ON DELETE CASCADE

	sqlProvider := `DELETE FROM providers WHERE id = $1`
	result, err := pps.db.Exec(ctx, sqlProvider, id)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return models.ErrProviderNotFound
	}

	return nil
}

// UpdateProviderStatus updates the status of a specific provider.
func (pps *PostgresProviderStore) UpdateProviderStatus(ctx context.Context, id uuid.UUID, status models.ProviderStatus) error {
	now := time.Now().UTC()
	sql := `
	UPDATE providers 
	SET status = $1, last_seen_at = $2
	WHERE id = $3
	`

	result, err := pps.db.Exec(ctx, sql, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update provider status: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return models.ErrProviderNotFound
	}

	return nil
}

// UpdateProviderHeartbeat updates the timestamp for the last heartbeat
// and also updates GPU utilization metrics if provided
func (pps *PostgresProviderStore) UpdateProviderHeartbeat(ctx context.Context, id uuid.UUID, gpuMetrics []models.GPUDetail) error {
	// Start a transaction
	tx, err := pps.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Update the provider's last_seen_at timestamp
	sql := `
	UPDATE providers 
	SET last_seen_at = NOW(), 
	    status = CASE WHEN status = 'offline' OR status = 'error' THEN 'idle' ELSE status END
	WHERE id = $1
	`
	result, err := tx.Exec(ctx, sql, id)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to update provider heartbeat: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback(ctx)
		return models.ErrProviderNotFound
	}

	// If GPU metrics are provided, update them
	if len(gpuMetrics) > 0 {
		// For each GPU, update its metrics
		for i, gpu := range gpuMetrics {
			// We need the GPU ID from the database
			var gpuID int
			err := tx.QueryRow(ctx,
				"SELECT id FROM gpu_details WHERE provider_id = $1 ORDER BY id LIMIT 1 OFFSET $2",
				id, i).Scan(&gpuID)

			if err != nil {
				tx.Rollback(ctx)
				if err == pgx.ErrNoRows {
					return fmt.Errorf("GPU at index %d not found for provider %s", i, id)
				}
				return fmt.Errorf("failed to get GPU ID: %w", err)
			}

			// Update the GPU metrics
			_, err = tx.Exec(ctx, `
				UPDATE gpu_details
				SET utilization_gpu_percent = $1,
					utilization_memory_percent = $2,
					temperature_c = $3,
					power_draw_w = $4,
					is_healthy = $5,
					updated_at = NOW()
				WHERE id = $6`,
				gpu.UtilizationGPU,
				gpu.UtilizationMem,
				gpu.Temperature,
				gpu.PowerDraw,
				gpu.IsHealthy,
				gpuID)

			if err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("failed to update GPU metrics: %w", err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection pool.
func (pps *PostgresProviderStore) Close() error {
	if pps.db != nil {
		pps.db.Close()
		pps.logger.Info("Closed PostgreSQL provider store connection pool")
	}
	return nil
}
