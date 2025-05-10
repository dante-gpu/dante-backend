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
	// Create providers table
	createProvidersTable := `
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
		metadata JSONB
	);

	-- Create indexes for frequently queried columns
	CREATE INDEX IF NOT EXISTS idx_providers_status ON providers(status);
	CREATE INDEX IF NOT EXISTS idx_providers_location ON providers(location);
	CREATE INDEX IF NOT EXISTS idx_providers_last_seen_at ON providers(last_seen_at);
	CREATE INDEX IF NOT EXISTS idx_providers_owner_id ON providers(owner_id);
	`

	// Create GPU details table
	createGPUDetailsTable := `
	CREATE TABLE IF NOT EXISTS gpu_details (
		id SERIAL PRIMARY KEY,
		provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
		model_name TEXT NOT NULL,
		vram_mb BIGINT NOT NULL,
		driver_version TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- Create index for the provider_id foreign key
	CREATE INDEX IF NOT EXISTS idx_gpu_details_provider_id ON gpu_details(provider_id);
	CREATE INDEX IF NOT EXISTS idx_gpu_details_model_name ON gpu_details(model_name);
	`

	// Execute the table creation queries
	if _, err := pps.db.Exec(ctx, createProvidersTable); err != nil {
		return fmt.Errorf("failed to create providers table: %w", err)
	}

	if _, err := pps.db.Exec(ctx, createGPUDetailsTable); err != nil {
		return fmt.Errorf("failed to create gpu_details table: %w", err)
	}

	pps.logger.Info("PostgreSQL tables initialized for provider store")
	return nil
}

// AddProvider adds a new provider to the PostgreSQL database.
func (pps *PostgresProviderStore) AddProvider(ctx context.Context, provider *models.Provider) error {
	// Start a transaction
	tx, err := pps.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	// Insert the provider
	metadataJSON, err := json.Marshal(provider.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	sqlProvider := `
	INSERT INTO providers (
		id, owner_id, name, hostname, ip_address, status, location, registered_at, last_seen_at, metadata
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = tx.Exec(ctx, sqlProvider,
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
		return fmt.Errorf("failed to insert provider: %w", err)
	}

	// Insert each GPU detail
	sqlGPU := `
	INSERT INTO gpu_details (
		provider_id, model_name, vram_mb, driver_version
	) VALUES ($1, $2, $3, $4)
	`

	for _, gpu := range provider.GPUs {
		_, err = tx.Exec(ctx, sqlGPU,
			provider.ID,
			gpu.ModelName,
			gpu.VRAM,
			gpu.DriverVersion,
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
	SELECT model_name, vram_mb, driver_version
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
		if err := rows.Scan(&gpu.ModelName, &gpu.VRAM, &gpu.DriverVersion); err != nil {
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
	// Query for providers and their GPU details in a single operation
	// using a LEFT JOIN and JSON aggregation to avoid N+1 problem
	sqlQuery := `
	SELECT 
		p.id, p.owner_id, p.name, p.hostname, p.ip_address, p.status, p.location, 
		p.registered_at, p.last_seen_at, p.metadata,
		COALESCE(
			json_agg(
				json_build_object(
					'model_name', g.model_name,
					'vram_mb', g.vram_mb,
					'driver_version', g.driver_version
				)
			) FILTER (WHERE g.id IS NOT NULL),
			'[]'::json
		) as gpus
	FROM providers p
	LEFT JOIN gpu_details g ON p.id = g.provider_id
	GROUP BY p.id
	`

	rows, err := pps.db.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query providers: %w", err)
	}
	defer rows.Close()

	providers := make([]*models.Provider, 0)

	for rows.Next() {
		var provider models.Provider
		var metadataJSON, gpusJSON []byte
		var hostname, ipAddress, location sql.NullString // For handling nullable fields

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
			return nil, fmt.Errorf("failed to scan provider: %w", err)
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
		if len(metadataJSON) > 0 && !strings.EqualFold(string(metadataJSON), "null") {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata for provider %s: %w", provider.ID, err)
			}
			provider.Metadata = metadata
		}

		// Unmarshal GPU details
		if len(gpusJSON) > 0 && !strings.EqualFold(string(gpusJSON), "null") && !strings.EqualFold(string(gpusJSON), "[]") {
			var gpus []models.GPUDetail
			if err := json.Unmarshal(gpusJSON, &gpus); err != nil {
				return nil, fmt.Errorf("failed to unmarshal GPU details for provider %s: %w", provider.ID, err)
			}
			provider.GPUs = gpus
		} else {
			provider.GPUs = []models.GPUDetail{} // Empty slice instead of nil
		}

		providers = append(providers, &provider)
	}

	if err = rows.Err(); err != nil {
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
		provider_id, model_name, vram_mb, driver_version
	) VALUES ($1, $2, $3, $4)
	`

	for _, gpu := range updatedProvider.GPUs {
		_, err = tx.Exec(ctx, sqlGPU,
			id,
			gpu.ModelName,
			gpu.VRAM,
			gpu.DriverVersion,
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

// UpdateProviderHeartbeat updates the LastSeenAt timestamp for a provider.
func (pps *PostgresProviderStore) UpdateProviderHeartbeat(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	// Also update status to idle if it was offline or error
	sql := `
	UPDATE providers 
	SET last_seen_at = $1,
		status = CASE 
			WHEN status = 'offline' OR status = 'error' THEN 'idle'::text 
			ELSE status 
		END
	WHERE id = $2
	`

	result, err := pps.db.Exec(ctx, sql, now, id)
	if err != nil {
		return fmt.Errorf("failed to update provider heartbeat: %w", err)
	}

	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return models.ErrProviderNotFound
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
