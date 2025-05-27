package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/dante-gpu/dante-backend/billing-payment-service/internal/models"
)

// PostgresStore implements billing data storage using PostgreSQL
type PostgresStore struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(db *pgxpool.Pool, logger *zap.Logger) *PostgresStore {
	return &PostgresStore{
		db:     db,
		logger: logger,
	}
}

// Initialize creates the necessary database tables
func (s *PostgresStore) Initialize(ctx context.Context) error {
	queries := []string{
		createWalletsTable,
		createTransactionsTable,
		createRentalSessionsTable,
		createUsageRecordsTable,
		createBillingRecordsTable,
		createProviderRatesTable,
		createIndexes,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	s.logger.Info("Database tables initialized successfully")
	return nil
}

// Wallet operations

// CreateWallet creates a new wallet
func (s *PostgresStore) CreateWallet(ctx context.Context, req *models.WalletCreateRequest) (*models.Wallet, error) {
	wallet := &models.Wallet{
		ID:             uuid.New(),
		UserID:         req.UserID,
		WalletType:     req.WalletType,
		SolanaAddress:  req.SolanaAddress,
		Balance:        decimal.Zero,
		LockedBalance:  decimal.Zero,
		PendingBalance: decimal.Zero,
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	query := `
		INSERT INTO wallets (id, user_id, wallet_type, solana_address, balance, locked_balance, pending_balance, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(ctx, query,
		wallet.ID, wallet.UserID, wallet.WalletType, wallet.SolanaAddress,
		wallet.Balance, wallet.LockedBalance, wallet.PendingBalance,
		wallet.IsActive, wallet.CreatedAt, wallet.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	s.logger.Info("Wallet created", zap.String("wallet_id", wallet.ID.String()), zap.String("user_id", wallet.UserID))
	return wallet, nil
}

// GetWallet retrieves a wallet by ID
func (s *PostgresStore) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `
		SELECT id, user_id, wallet_type, solana_address, balance, locked_balance, pending_balance, 
		       is_active, created_at, updated_at, last_activity_at
		FROM wallets WHERE id = $1
	`

	var lastActivityAt sql.NullTime
	err := s.db.QueryRow(ctx, query, walletID).Scan(
		&wallet.ID, &wallet.UserID, &wallet.WalletType, &wallet.SolanaAddress,
		&wallet.Balance, &wallet.LockedBalance, &wallet.PendingBalance,
		&wallet.IsActive, &wallet.CreatedAt, &wallet.UpdatedAt, &lastActivityAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	if lastActivityAt.Valid {
		wallet.LastActivityAt = &lastActivityAt.Time
	}

	return wallet, nil
}

// GetWalletByUserID retrieves a wallet by user ID and type
func (s *PostgresStore) GetWalletByUserID(ctx context.Context, userID string, walletType models.WalletType) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `
		SELECT id, user_id, wallet_type, solana_address, balance, locked_balance, pending_balance, 
		       is_active, created_at, updated_at, last_activity_at
		FROM wallets WHERE user_id = $1 AND wallet_type = $2
	`

	var lastActivityAt sql.NullTime
	err := s.db.QueryRow(ctx, query, userID, walletType).Scan(
		&wallet.ID, &wallet.UserID, &wallet.WalletType, &wallet.SolanaAddress,
		&wallet.Balance, &wallet.LockedBalance, &wallet.PendingBalance,
		&wallet.IsActive, &wallet.CreatedAt, &wallet.UpdatedAt, &lastActivityAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	if lastActivityAt.Valid {
		wallet.LastActivityAt = &lastActivityAt.Time
	}

	return wallet, nil
}

// UpdateWalletBalance updates wallet balance and locked balance
func (s *PostgresStore) UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, balance, lockedBalance decimal.Decimal) error {
	query := `
		UPDATE wallets 
		SET balance = $2, locked_balance = $3, updated_at = $4, last_activity_at = $4
		WHERE id = $1
	`

	result, err := s.db.Exec(ctx, query, walletID, balance, lockedBalance, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrWalletNotFound
	}

	return nil
}

// Transaction operations

// CreateTransaction creates a new transaction
func (s *PostgresStore) CreateTransaction(ctx context.Context, req *models.TransactionCreateRequest) (*models.Transaction, error) {
	transaction := &models.Transaction{
		ID:            uuid.New(),
		FromWalletID:  req.FromWalletID,
		ToWalletID:    req.ToWalletID,
		Type:          req.Type,
		Status:        models.TransactionStatusPending,
		Amount:        req.Amount,
		Fee:           decimal.Zero, // Will be calculated based on transaction type
		Description:   req.Description,
		SessionID:     req.SessionID,
		JobID:         req.JobID,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	metadataJSON, err := json.Marshal(transaction.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO transactions (id, from_wallet_id, to_wallet_id, type, status, amount, fee, description, 
		                         session_id, job_id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = s.db.Exec(ctx, query,
		transaction.ID, transaction.FromWalletID, transaction.ToWalletID,
		transaction.Type, transaction.Status, transaction.Amount, transaction.Fee,
		transaction.Description, transaction.SessionID, transaction.JobID,
		metadataJSON, transaction.CreatedAt, transaction.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	s.logger.Info("Transaction created",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("type", string(transaction.Type)),
		zap.String("amount", transaction.Amount.String()),
	)

	return transaction, nil
}

// UpdateTransactionStatus updates transaction status and signature
func (s *PostgresStore) UpdateTransactionStatus(ctx context.Context, transactionID uuid.UUID, status models.TransactionStatus, signature *string) error {
	var confirmedAt *time.Time
	if status == models.TransactionStatusConfirmed {
		now := time.Now().UTC()
		confirmedAt = &now
	}

	query := `
		UPDATE transactions 
		SET status = $2, solana_signature = $3, confirmed_at = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := s.db.Exec(ctx, query, transactionID, status, signature, confirmedAt, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrTransactionNotFound
	}

	return nil
}

// GetTransaction retrieves a transaction by ID
func (s *PostgresStore) GetTransaction(ctx context.Context, transactionID uuid.UUID) (*models.Transaction, error) {
	transaction := &models.Transaction{}
	query := `
		SELECT id, from_wallet_id, to_wallet_id, type, status, amount, fee, description,
		       solana_signature, session_id, job_id, metadata, created_at, updated_at, confirmed_at
		FROM transactions WHERE id = $1
	`

	var metadataJSON []byte
	var confirmedAt sql.NullTime
	err := s.db.QueryRow(ctx, query, transactionID).Scan(
		&transaction.ID, &transaction.FromWalletID, &transaction.ToWalletID,
		&transaction.Type, &transaction.Status, &transaction.Amount, &transaction.Fee,
		&transaction.Description, &transaction.SolanaSignature, &transaction.SessionID,
		&transaction.JobID, &metadataJSON, &transaction.CreatedAt, &transaction.UpdatedAt,
		&confirmedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if confirmedAt.Valid {
		transaction.ConfirmedAt = &confirmedAt.Time
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &transaction.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return transaction, nil
}
