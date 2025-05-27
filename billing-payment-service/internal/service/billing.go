package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/dante-gpu/dante-backend/billing-payment-service/internal/models"
	"github.com/dante-gpu/dante-backend/billing-payment-service/internal/pricing"
	"github.com/dante-gpu/dante-backend/billing-payment-service/internal/solana"
	"github.com/dante-gpu/dante-backend/billing-payment-service/internal/store"
)

// BillingService handles all billing and payment operations
type BillingService struct {
	store         *store.PostgresStore
	solanaClient  *solana.Client
	pricingEngine *pricing.Engine
	logger        *zap.Logger
	config        *Config
}

// Config represents billing service configuration
type Config struct {
	MinimumBalance          decimal.Decimal `yaml:"minimum_balance"`
	LowBalanceThreshold     decimal.Decimal `yaml:"low_balance_threshold"`
	BillingInterval         time.Duration   `yaml:"billing_interval"`
	InsufficientFundsGrace  time.Duration   `yaml:"insufficient_funds_grace_period"`
	MaxTransactionAmount    decimal.Decimal `yaml:"max_transaction_amount"`
	DailyWithdrawalLimit    decimal.Decimal `yaml:"daily_withdrawal_limit"`
	MinimumPayoutAmount     decimal.Decimal `yaml:"minimum_payout_amount"`
	PayoutFeePercent        decimal.Decimal `yaml:"payout_fee_percent"`
}

// NewBillingService creates a new billing service
func NewBillingService(
	store *store.PostgresStore,
	solanaClient *solana.Client,
	pricingEngine *pricing.Engine,
	config *Config,
	logger *zap.Logger,
) *BillingService {
	return &BillingService{
		store:         store,
		solanaClient:  solanaClient,
		pricingEngine: pricingEngine,
		config:        config,
		logger:        logger,
	}
}

// Wallet Management

// CreateWallet creates a new dGPU token wallet for a user or provider
func (s *BillingService) CreateWallet(ctx context.Context, req *models.WalletCreateRequest) (*models.Wallet, error) {
	s.logger.Info("Creating wallet",
		zap.String("user_id", req.UserID),
		zap.String("wallet_type", string(req.WalletType)),
		zap.String("solana_address", req.SolanaAddress),
	)

	// Validate Solana address format
	if !s.isValidSolanaAddress(req.SolanaAddress) {
		return nil, models.NewValidationError("solana_address", "invalid Solana address format")
	}

	// Check if wallet already exists for this user and type
	existingWallet, err := s.store.GetWalletByUserID(ctx, req.UserID, req.WalletType)
	if err != nil && err != models.ErrWalletNotFound {
		return nil, fmt.Errorf("failed to check existing wallet: %w", err)
	}
	if existingWallet != nil {
		return nil, models.NewBillingError(models.ErrCodeWalletExists, "Wallet already exists", models.ErrWalletAlreadyExists)
	}

	// Create associated token account on Solana if needed
	_, err = s.solanaClient.CreateAssociatedTokenAccount(ctx, req.SolanaAddress)
	if err != nil {
		s.logger.Error("Failed to create associated token account", zap.Error(err))
		return nil, models.NewSolanaError("create_ata", err)
	}

	// Create wallet in database
	wallet, err := s.store.CreateWallet(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	s.logger.Info("Wallet created successfully", zap.String("wallet_id", wallet.ID.String()))
	return wallet, nil
}

// GetWalletBalance gets the current balance of a wallet
func (s *BillingService) GetWalletBalance(ctx context.Context, walletID uuid.UUID) (*models.BalanceResponse, error) {
	wallet, err := s.store.GetWallet(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Get real-time balance from Solana
	solanaBalance, err := s.solanaClient.GetTokenBalance(ctx, wallet.SolanaAddress)
	if err != nil {
		s.logger.Warn("Failed to get Solana balance, using database balance", zap.Error(err))
		solanaBalance = wallet.Balance
	}

	// Update database balance if there's a significant difference
	if solanaBalance.Sub(wallet.Balance).Abs().GreaterThan(decimal.NewFromFloat(0.001)) {
		err = s.store.UpdateWalletBalance(ctx, walletID, solanaBalance, wallet.LockedBalance)
		if err != nil {
			s.logger.Warn("Failed to update wallet balance", zap.Error(err))
		} else {
			wallet.Balance = solanaBalance
		}
	}

	return &models.BalanceResponse{
		WalletID:         wallet.ID,
		Balance:          wallet.Balance,
		LockedBalance:    wallet.LockedBalance,
		PendingBalance:   wallet.PendingBalance,
		AvailableBalance: wallet.AvailableBalance(),
		TotalBalance:     wallet.TotalBalance(),
		LastUpdated:      wallet.UpdatedAt,
	}, nil
}

// Session Management

// StartRentalSession starts a new GPU rental session
func (s *BillingService) StartRentalSession(ctx context.Context, req *models.SessionStartRequest) (*models.SessionResponse, error) {
	s.logger.Info("Starting rental session",
		zap.String("user_id", req.UserID),
		zap.String("provider_id", req.ProviderID.String()),
		zap.String("gpu_model", req.GPUModel),
		zap.Uint64("requested_vram_mb", req.RequestedVRAM),
	)

	// Get user wallet
	userWallet, err := s.store.GetWalletByUserID(ctx, req.UserID, models.WalletTypeUser)
	if err != nil {
		return nil, err
	}

	// Check minimum balance
	if userWallet.AvailableBalance().LessThan(s.config.MinimumBalance) {
		return nil, models.NewInsufficientFundsError(
			s.config.MinimumBalance.String(),
			userWallet.AvailableBalance().String(),
		)
	}

	// Calculate pricing for initial hour
	pricingReq := &pricing.PricingRequest{
		GPUModel:        req.GPUModel,
		RequestedVRAM:   req.RequestedVRAM,
		TotalVRAM:       req.RequestedVRAM, // This should come from provider registry
		EstimatedPowerW: req.EstimatedPowerW,
		DurationHours:   decimal.NewFromInt(1),
		ProviderID:      &req.ProviderID,
		UserID:          &req.UserID,
	}

	pricing, err := s.pricingEngine.CalculatePricing(ctx, pricingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate pricing: %w", err)
	}

	// Check if user can afford at least one hour
	if userWallet.AvailableBalance().LessThan(pricing.TotalHourlyCost) {
		return nil, models.NewInsufficientFundsError(
			pricing.TotalHourlyCost.String(),
			userWallet.AvailableBalance().String(),
		)
	}

	// Lock funds for initial hour
	err = userWallet.LockFunds(pricing.TotalHourlyCost)
	if err != nil {
		return nil, err
	}

	// Update wallet in database
	err = s.store.UpdateWalletBalance(ctx, userWallet.ID, userWallet.Balance, userWallet.LockedBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to lock funds: %w", err)
	}

	// Create rental session
	session := &models.RentalSession{
		ID:                uuid.New(),
		UserID:            req.UserID,
		ProviderID:        req.ProviderID,
		JobID:             req.JobID,
		Status:            models.SessionStatusActive,
		GPUModel:          req.GPUModel,
		AllocatedVRAM:     req.RequestedVRAM,
		TotalVRAM:         req.RequestedVRAM, // This should come from provider registry
		VRAMPercentage:    decimal.NewFromInt(100), // Assuming full allocation for now
		HourlyRate:        pricing.BaseHourlyRate,
		VRAMRate:          pricing.VRAMHourlyRate,
		PowerRate:         pricing.PowerHourlyRate,
		PlatformFeeRate:   decimal.NewFromFloat(5.0), // From config
		EstimatedPowerW:   req.EstimatedPowerW,
		StartedAt:         time.Now().UTC(),
		LastBilledAt:      time.Now().UTC(),
		TotalCost:         decimal.Zero,
		PlatformFee:       decimal.Zero,
		ProviderEarnings:  decimal.Zero,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// Save session to database (this would need to be implemented in store)
	// err = s.store.CreateRentalSession(ctx, session)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create rental session: %w", err)
	// }

	// Create initial transaction record
	txnReq := &models.TransactionCreateRequest{
		FromWalletID: &userWallet.ID,
		Type:         models.TransactionTypeSessionStart,
		Amount:       pricing.TotalHourlyCost,
		Description:  fmt.Sprintf("Session start - locked funds for %s", req.GPUModel),
		SessionID:    &session.ID,
		JobID:        req.JobID,
	}

	_, err = s.store.CreateTransaction(ctx, txnReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Calculate estimated runtime based on available balance
	estimatedRuntime := userWallet.AvailableBalance().Div(pricing.TotalHourlyCost)

	response := &models.SessionResponse{
		Session:              *session,
		CurrentCost:          decimal.Zero,
		EstimatedHourlyCost:  pricing.TotalHourlyCost,
		RemainingBalance:     userWallet.AvailableBalance(),
		EstimatedRuntime:     estimatedRuntime,
	}

	s.logger.Info("Rental session started successfully",
		zap.String("session_id", session.ID.String()),
		zap.String("estimated_runtime_hours", estimatedRuntime.String()),
	)

	return response, nil
}

// ProcessUsageUpdate processes real-time usage data from provider daemon
func (s *BillingService) ProcessUsageUpdate(ctx context.Context, req *models.UsageUpdateRequest) error {
	s.logger.Debug("Processing usage update",
		zap.String("session_id", req.SessionID.String()),
		zap.Uint8("gpu_utilization", req.GPUUtilization),
		zap.Uint32("power_draw", req.PowerDraw),
	)

	// Get session (this would need to be implemented in store)
	// session, err := s.store.GetRentalSession(ctx, req.SessionID)
	// if err != nil {
	//     return err
	// }

	// Create usage record
	usageRecord := &models.UsageRecord{
		ID:                   uuid.New(),
		SessionID:            req.SessionID,
		RecordedAt:           req.Timestamp,
		GPUUtilization:       req.GPUUtilization,
		VRAMUtilization:      req.VRAMUtilization,
		PowerDraw:            req.PowerDraw,
		Temperature:          req.Temperature,
		PeriodMinutes:        1, // Assuming 1-minute intervals
		PeriodCost:           decimal.Zero, // Will be calculated
		CreatedAt:            time.Now().UTC(),
	}

	// Save usage record (this would need to be implemented in store)
	// err = s.store.CreateUsageRecord(ctx, usageRecord)
	// if err != nil {
	//     return fmt.Errorf("failed to create usage record: %w", err)
	// }

	s.logger.Debug("Usage update processed successfully")
	return nil
}

// Helper methods

// isValidSolanaAddress validates a Solana address format
func (s *BillingService) isValidSolanaAddress(address string) bool {
	// Basic validation - in production, use proper Solana address validation
	return len(address) >= 32 && len(address) <= 44
}
