package store

import (
	"context"
	"encoding/json" // For metadata if we decide to marshal/unmarshal manually in some cases
	"fmt"
	"strings" // For errors.Is check with older pgx versions or specific error strings
	"time"    // For timestamps

	"github.com/dante-gpu/dante-backend/provider-registry-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5" // Import pgx for pgx.ErrNoRows
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgresProviderStore implements the ProviderStore interface using a PostgreSQL database.
type PostgresProviderStore struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgresProviderStore creates a new PostgresProviderStore.
// It expects a connected pgxpool.Pool.
func NewPostgresProviderStore(db *pgxpool.Pool, logger *zap.Logger) *PostgresProviderStore {
	return &PostgresProviderStore{
		db:     db,
		logger: logger,
	}
}

// Initialize creates the necessary 'providers' and 'gpu_details' tables if they don't already exist.
func (pps *PostgresProviderStore) Initialize(ctx context.Context) error {
	pps.logger.Info("Initializing PostgreSQL schema for provider registry...")

	createProvidersTableSQL := `
	CREATE TABLE IF NOT EXISTS providers (
		id UUID PRIMARY KEY,
		owner_id VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		hostname VARCHAR(255),
		ip_address VARCHAR(255),
		status VARCHAR(50) NOT NULL,
		location VARCHAR(255),
		registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		metadata JSONB
	);
	`

	_, err := pps.db.Exec(ctx, createProvidersTableSQL)
	if err != nil {
		pps.logger.Error("Failed to create 'providers' table", zap.Error(err))
		return fmt.Errorf("initializing providers table: %w", err)
	}
	pps.logger.Info("'providers' table checked/created successfully")

	createGpuDetailsTableSQL := `
	CREATE TABLE IF NOT EXISTS gpu_details (
		id SERIAL PRIMARY KEY,
		provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
		model_name VARCHAR(255) NOT NULL,
		vram_mb BIGINT NOT NULL,
		driver_version VARCHAR(100) NOT NULL
	);
	`

	_, err = pps.db.Exec(ctx, createGpuDetailsTableSQL)
	if err != nil {
		pps.logger.Error("Failed to create 'gpu_details' table", zap.Error(err))
		return fmt.Errorf("initializing gpu_details table: %w", err)
	}
	pps.logger.Info("'gpu_details' table checked/created successfully")

	// Create indexes for frequently queried columns
	createIndexesSQL := `
	CREATE INDEX IF NOT EXISTS idx_providers_status ON providers (status);
	CREATE INDEX IF NOT EXISTS idx_providers_location ON providers (location);
	CREATE INDEX IF NOT EXISTS idx_providers_last_seen_at ON providers (last_seen_at DESC);
	CREATE INDEX IF NOT EXISTS idx_gpu_details_provider_id ON gpu_details (provider_id);
	CREATE INDEX IF NOT EXISTS idx_gpu_details_model_name ON gpu_details (model_name);
	`
	_, err = pps.db.Exec(ctx, createIndexesSQL)
	if err != nil {
		pps.logger.Error("Failed to create indexes for tables", zap.Error(err))
		return fmt.Errorf("creating indexes: %w", err)
	}
	pps.logger.Info("Indexes checked/created successfully")

	pps.logger.Info("PostgreSQL schema initialization complete.")
	return nil
}

// AddProvider adds a new provider and its GPU details to the database within a transaction.
func (pps *PostgresProviderStore) AddProvider(provider *models.Provider) error {
	ctx := context.Background()

	tx, err := pps.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction for AddProvider: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if anything goes wrong

	addProviderSQL := `
	INSERT INTO providers (id, owner_id, name, hostname, ip_address, status, location, registered_at, last_seen_at, metadata)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.Exec(ctx, addProviderSQL,
		provider.ID,
		provider.OwnerID,
		provider.Name,
		provider.Hostname,
		provider.IPAddress,
		provider.Status,
		provider.Location,
		provider.RegisteredAt,
		provider.LastSeenAt,
		provider.Metadata,
	)
	if err != nil {
		pps.logger.Error("Failed to insert into 'providers' table", zap.String("provider_id", provider.ID.String()), zap.Error(err))
		return fmt.Errorf("inserting provider %s: %w", provider.ID.String(), err)
	}

	addGpuDetailSQL := `
	INSERT INTO gpu_details (provider_id, model_name, vram_mb, driver_version)
	VALUES ($1, $2, $3, $4)
	`
	for _, gpu := range provider.GPUs {
		_, err = tx.Exec(ctx, addGpuDetailSQL, provider.ID, gpu.ModelName, gpu.VRAM, gpu.DriverVersion)
		if err != nil {
			pps.logger.Error("Failed to insert into 'gpu_details' table",
				zap.String("provider_id", provider.ID.String()),
				zap.String("gpu_model", gpu.ModelName),
				zap.Error(err),
			)
			return fmt.Errorf("inserting GPU detail for provider %s: %w", provider.ID.String(), err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		pps.logger.Error("Failed to commit transaction for AddProvider", zap.String("provider_id", provider.ID.String()), zap.Error(err))
		return fmt.Errorf("committing transaction for AddProvider: %w", err)
	}

	pps.logger.Info("Successfully added provider and GPU details to DB", zap.String("provider_id", provider.ID.String()))
	return nil
}

// GetProvider retrieves a provider and its GPU details by ID.
func (pps *PostgresProviderStore) GetProvider(id uuid.UUID) (*models.Provider, error) {
	ctx := context.Background()
	provider := &models.Provider{}

	getProviderSQL := `
	SELECT id, owner_id, name, hostname, ip_address, status, location, registered_at, last_seen_at, metadata
	FROM providers
	WHERE id = $1
	`
	var metadataBytes []byte

	err := pps.db.QueryRow(ctx, getProviderSQL, id).Scan(
		&provider.ID,
		&provider.OwnerID,
		&provider.Name,
		&provider.Hostname,
		&provider.IPAddress,
		&provider.Status,
		&provider.Location,
		&provider.RegisteredAt,
		&provider.LastSeenAt,
		&metadataBytes,
	)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") || err == pgx.ErrNoRows {
			pps.logger.Debug("Provider not found in DB for GetProvider", zap.String("provider_id", id.String()))
			return nil, models.ErrProviderNotFound
		}
		pps.logger.Error("Failed to get provider from DB", zap.String("provider_id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("getting provider %s: %w", id.String(), err)
	}

	if len(metadataBytes) > 0 && string(metadataBytes) != "null" {
		if err := json.Unmarshal(metadataBytes, &provider.Metadata); err != nil {
			pps.logger.Error("Failed to unmarshal metadata for provider", zap.String("provider_id", id.String()), zap.Error(err))
			return nil, fmt.Errorf("unmarshalling metadata for provider %s: %w", id.String(), err)
		}
	} else {
		provider.Metadata = nil // Ensure it's nil if DB returned null or empty JSON
	}

	getGpusSQL := `
	SELECT model_name, vram_mb, driver_version
	FROM gpu_details
	WHERE provider_id = $1
	`
	rows, err := pps.db.Query(ctx, getGpusSQL, id)
	if err != nil {
		pps.logger.Error("Failed to get GPU details for provider from DB", zap.String("provider_id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("getting GPU details for provider %s: %w", id.String(), err)
	}
	defer rows.Close()

	var gpus []models.GPUDetail
	for rows.Next() {
		gpu := models.GPUDetail{}
		if err := rows.Scan(&gpu.ModelName, &gpu.VRAM, &gpu.DriverVersion); err != nil {
			pps.logger.Error("Failed to scan GPU detail row", zap.String("provider_id", id.String()), zap.Error(err))
			return nil, fmt.Errorf("scanning GPU detail for provider %s: %w", id.String(), err)
		}
		gpus = append(gpus, gpu)
	}
	if err := rows.Err(); err != nil {
		pps.logger.Error("Error iterating over GPU detail rows", zap.String("provider_id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("iterating GPU details for provider %s: %w", id.String(), err)
	}
	provider.GPUs = gpus

	pps.logger.Debug("Successfully retrieved provider and GPU details from DB", zap.String("provider_id", id.String()))
	return provider, nil
}

// ListProviders retrieves a list of providers.
// TODO: Implement comprehensive filtering and pagination.
func (pps *PostgresProviderStore) ListProviders() ([]*models.Provider, error) {
	ctx := context.Background()
	var providers []*models.Provider

	// Basic query, to be expanded with filters and pagination
	listProvidersSQL := `
    SELECT p.id, p.owner_id, p.name, p.hostname, p.ip_address, p.status, p.location, p.registered_at, p.last_seen_at, p.metadata,
           COALESCE(json_agg(json_build_object('model_name', g.model_name, 'vram_mb', g.vram_mb, 'driver_version', g.driver_version)) FILTER (WHERE g.provider_id IS NOT NULL), '[]') AS gpus
    FROM providers p
    LEFT JOIN gpu_details g ON p.id = g.provider_id
    GROUP BY p.id
    ORDER BY p.registered_at DESC
    `
	// Note: The COALESCE and FILTER are to handle providers with no GPUs correctly, returning an empty JSON array '[]' for gpus.

	rows, err := pps.db.Query(ctx, listProvidersSQL)
	if err != nil {
		pps.logger.Error("Failed to list providers from DB", zap.Error(err))
		return nil, fmt.Errorf("listing providers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		provider := models.Provider{}
		var metadataBytes []byte
		var gpusBytes []byte // To scan the aggregated JSON for GPUs

		err := rows.Scan(
			&provider.ID,
			&provider.OwnerID,
			&provider.Name,
			&provider.Hostname,
			&provider.IPAddress,
			&provider.Status,
			&provider.Location,
			&provider.RegisteredAt,
			&provider.LastSeenAt,
			&metadataBytes,
			&gpusBytes,
		)
		if err != nil {
			pps.logger.Error("Failed to scan provider row for ListProviders", zap.Error(err))
			continue // Skip problematic row
		}

		if len(metadataBytes) > 0 && string(metadataBytes) != "null" {
			if err := json.Unmarshal(metadataBytes, &provider.Metadata); err != nil {
				pps.logger.Warn("Failed to unmarshal metadata for provider in ListProviders", zap.String("provider_id", provider.ID.String()), zap.Error(err))
				provider.Metadata = nil
			}
		}

		if len(gpusBytes) > 0 {
			if err := json.Unmarshal(gpusBytes, &provider.GPUs); err != nil {
				pps.logger.Warn("Failed to unmarshal GPUs for provider in ListProviders", zap.String("provider_id", provider.ID.String()), zap.Error(err))
				provider.GPUs = []models.GPUDetail{} // Ensure it's an empty slice, not nil
			}
		} else {
			provider.GPUs = []models.GPUDetail{} // Ensure it's an empty slice if no GPUs
		}

		providers = append(providers, &provider)
	}
	if err := rows.Err(); err != nil {
		pps.logger.Error("Error iterating over provider rows in ListProviders", zap.Error(err))
		return nil, fmt.Errorf("iterating provider rows for ListProviders: %w", err)
	}

	pps.logger.Debug("Successfully listed providers from DB", zap.Int("count", len(providers)))
	return providers, nil
}

// UpdateProvider updates an existing provider's details.
// TODO: Handle GPU details update (e.g., delete existing and insert new ones, or a more sophisticated diff).
func (pps *PostgresProviderStore) UpdateProvider(id uuid.UUID, updatedProvider *models.Provider) error {
	ctx := context.Background()
	updatedProvider.LastSeenAt = time.Now().UTC()

	// For a full update, it's often easier to manage with a transaction if GPUs are also updated.
	tx, err := pps.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction for UpdateProvider: %w", err)
	}
	defer tx.Rollback(ctx)

	updateSQL := `
	UPDATE providers
	SET owner_id = $1, name = $2, hostname = $3, ip_address = $4, status = $5, location = $6, last_seen_at = $7, metadata = $8
	WHERE id = $9
	`
	cmdTag, err := tx.Exec(ctx, updateSQL,
		updatedProvider.OwnerID,
		updatedProvider.Name,
		updatedProvider.Hostname,
		updatedProvider.IPAddress,
		updatedProvider.Status,
		updatedProvider.Location,
		updatedProvider.LastSeenAt,
		updatedProvider.Metadata,
		id,
	)
	if err != nil {
		pps.logger.Error("Failed to update provider in DB", zap.String("provider_id", id.String()), zap.Error(err))
		return fmt.Errorf("updating provider %s: %w", id.String(), err)
	}
	if cmdTag.RowsAffected() == 0 {
		return models.ErrProviderNotFound
	}

	// GPU Management for UpdateProvider:
	// 1. Delete existing GPU details for this provider.
	// 2. Insert new GPU details from updatedProvider.GPUs.
	deleteGpusSQL := `DELETE FROM gpu_details WHERE provider_id = $1`
	_, err = tx.Exec(ctx, deleteGpusSQL, id)
	if err != nil {
		pps.logger.Error("Failed to delete old GPU details during update", zap.String("provider_id", id.String()), zap.Error(err))
		return fmt.Errorf("deleting old gpus for provider %s: %w", id.String(), err)
	}

	addGpuDetailSQL := `
	INSERT INTO gpu_details (provider_id, model_name, vram_mb, driver_version)
	VALUES ($1, $2, $3, $4)
	`
	for _, gpu := range updatedProvider.GPUs {
		_, err = tx.Exec(ctx, addGpuDetailSQL, id, gpu.ModelName, gpu.VRAM, gpu.DriverVersion)
		if err != nil {
			pps.logger.Error("Failed to insert new GPU detail during update",
				zap.String("provider_id", id.String()),
				zap.String("gpu_model", gpu.ModelName),
				zap.Error(err),
			)
			return fmt.Errorf("inserting new GPU detail for provider %s: %w", id.String(), err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		pps.logger.Error("Failed to commit transaction for UpdateProvider", zap.String("provider_id", id.String()), zap.Error(err))
		return fmt.Errorf("committing transaction for UpdateProvider: %w", err)
	}

	pps.logger.Info("Successfully updated provider and its GPU details in DB", zap.String("provider_id", id.String()))
	return nil
}

// DeleteProvider removes a provider and its associated GPU details (due to ON DELETE CASCADE).
func (pps *PostgresProviderStore) DeleteProvider(id uuid.UUID) error {
	ctx := context.Background()
	deleteSQL := `DELETE FROM providers WHERE id = $1`
	cmdTag, err := pps.db.Exec(ctx, deleteSQL, id)
	if err != nil {
		pps.logger.Error("Failed to delete provider from DB", zap.String("provider_id", id.String()), zap.Error(err))
		return fmt.Errorf("deleting provider %s: %w", id.String(), err)
	}
	if cmdTag.RowsAffected() == 0 {
		return models.ErrProviderNotFound
	}
	pps.logger.Info("Successfully deleted provider from DB", zap.String("provider_id", id.String()))
	return nil
}

// UpdateProviderStatus updates only the status and last_seen_at timestamp of a provider.
func (pps *PostgresProviderStore) UpdateProviderStatus(id uuid.UUID, status models.ProviderStatus) error {
	ctx := context.Background()
	updateStatusSQL := `
	UPDATE providers
	SET status = $1, last_seen_at = $2
	WHERE id = $3
	`
	lastSeenAt := time.Now().UTC()
	cmdTag, err := pps.db.Exec(ctx, updateStatusSQL, status, lastSeenAt, id)
	if err != nil {
		pps.logger.Error("Failed to update provider status in DB", zap.String("provider_id", id.String()), zap.Error(err))
		return fmt.Errorf("updating status for provider %s: %w", id.String(), err)
	}
	if cmdTag.RowsAffected() == 0 {
		return models.ErrProviderNotFound
	}
	pps.logger.Info("Successfully updated provider status in DB", zap.String("provider_id", id.String()), zap.String("new_status", string(status)))
	return nil
}

// UpdateProviderHeartbeat updates the last_seen_at timestamp and potentially sets status to idle.
func (pps *PostgresProviderStore) UpdateProviderHeartbeat(id uuid.UUID) error {
	ctx := context.Background()
	updateHeartbeatSQL := `
	UPDATE providers
	SET last_seen_at = $1,
	    status = CASE
	                 WHEN status = $2 OR status = $3 THEN $4
	                 ELSE status
	             END
	WHERE id = $5
	`
	lastSeenAt := time.Now().UTC()
	cmdTag, err := pps.db.Exec(ctx, updateHeartbeatSQL,
		lastSeenAt,
		models.StatusOffline,
		models.StatusError,
		models.StatusIdle,
		id,
	)
	if err != nil {
		pps.logger.Error("Failed to update provider heartbeat in DB", zap.String("provider_id", id.String()), zap.Error(err))
		return fmt.Errorf("updating heartbeat for provider %s: %w", id.String(), err)
	}
	if cmdTag.RowsAffected() == 0 {
		return models.ErrProviderNotFound
	}
	pps.logger.Info("Successfully updated provider heartbeat in DB", zap.String("provider_id", id.String()))
	return nil
}

// Close is a placeholder as pgxpool.Pool is managed externally.
func (pps *PostgresProviderStore) Close() error {
	pps.logger.Info("PostgresProviderStore Close called. Pool management is external.")
	return nil
}
