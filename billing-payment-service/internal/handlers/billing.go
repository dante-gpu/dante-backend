package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/dante-gpu/dante-backend/billing-payment-service/internal/models"
	"github.com/dante-gpu/dante-backend/billing-payment-service/internal/service"
)

// StartRentalSession handles rental session start requests
func StartRentalSession(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SessionStartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode session start request", zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		session, err := billingService.StartRentalSession(r.Context(), &req)
		if err != nil {
			logger.Error("Failed to start rental session", zap.Error(err))
			if billingErr, ok := err.(*models.BillingError); ok {
				writeErrorResponse(w, getHTTPStatusFromBillingError(billingErr), billingErr.Message, err)
			} else {
				writeErrorResponse(w, http.StatusInternalServerError, "Failed to start rental session", err)
			}
			return
		}

		logger.Info("Rental session started successfully",
			zap.String("session_id", session.Session.ID.String()),
			zap.String("user_id", session.Session.UserID),
			zap.String("provider_id", session.Session.ProviderID.String()),
		)

		writeJSONResponse(w, http.StatusCreated, session)
	}
}

// EndRentalSession handles rental session end requests
func EndRentalSession(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SessionEndRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode session end request", zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		// TODO: Implement session ending
		// session, err := billingService.EndRentalSession(r.Context(), &req)
		// if err != nil {
		//     logger.Error("Failed to end rental session", zap.Error(err))
		//     writeErrorResponse(w, http.StatusInternalServerError, "Failed to end rental session", err)
		//     return
		// }

		response := map[string]interface{}{
			"message":    "Session end request received",
			"session_id": req.SessionID,
			"reason":     req.Reason,
			"status":     "processing",
		}

		logger.Info("Rental session end requested",
			zap.String("session_id", req.SessionID.String()),
			zap.String("reason", req.Reason),
		)

		writeJSONResponse(w, http.StatusAccepted, response)
	}
}

// ProcessUsageUpdate handles real-time usage updates from provider daemons
func ProcessUsageUpdate(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.UsageUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode usage update request", zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		err := billingService.ProcessUsageUpdate(r.Context(), &req)
		if err != nil {
			logger.Error("Failed to process usage update", zap.Error(err))
			if billingErr, ok := err.(*models.BillingError); ok {
				writeErrorResponse(w, getHTTPStatusFromBillingError(billingErr), billingErr.Message, err)
			} else {
				writeErrorResponse(w, http.StatusInternalServerError, "Failed to process usage update", err)
			}
			return
		}

		logger.Debug("Usage update processed successfully",
			zap.String("session_id", req.SessionID.String()),
			zap.Uint8("gpu_utilization", req.GPUUtilization),
			zap.Uint32("power_draw", req.PowerDraw),
		)

		response := map[string]interface{}{
			"message": "Usage update processed successfully",
			"status":  "success",
		}

		writeJSONResponse(w, http.StatusOK, response)
	}
}

// GetCurrentUsage handles current usage requests for active sessions
func GetCurrentUsage(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionIDStr := chi.URLParam(r, "sessionID")
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			logger.Error("Invalid session ID", zap.String("session_id", sessionIDStr), zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid session ID", err)
			return
		}

		// TODO: Implement current usage retrieval
		// usage, err := billingService.GetCurrentUsage(r.Context(), sessionID)
		// if err != nil {
		//     logger.Error("Failed to get current usage", zap.String("session_id", sessionIDStr), zap.Error(err))
		//     writeErrorResponse(w, http.StatusInternalServerError, "Failed to get current usage", err)
		//     return
		// }

		// Placeholder response
		response := map[string]interface{}{
			"session_id":            sessionID,
			"current_cost":          "0.0",
			"estimated_hourly_cost": "1.5",
			"duration_minutes":      0,
			"gpu_utilization":       0,
			"vram_utilization":      0,
			"power_draw":            0,
			"status":                "active",
		}

		writeJSONResponse(w, http.StatusOK, response)
	}
}

// GetBillingHistory handles billing history requests
func GetBillingHistory(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &models.BillingHistoryRequest{
			Limit:  50, // Default limit
			Offset: 0,  // Default offset
		}

		// Parse query parameters
		if userID := r.URL.Query().Get("user_id"); userID != "" {
			req.UserID = &userID
		}

		if providerIDStr := r.URL.Query().Get("provider_id"); providerIDStr != "" {
			if providerID, err := uuid.Parse(providerIDStr); err == nil {
				req.ProviderID = &providerID
			}
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
				req.Limit = limit
			}
		}

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
				req.Offset = offset
			}
		}

		// TODO: Implement billing history retrieval
		// history, err := billingService.GetBillingHistory(r.Context(), req)
		// if err != nil {
		//     logger.Error("Failed to get billing history", zap.Error(err))
		//     writeErrorResponse(w, http.StatusInternalServerError, "Failed to get billing history", err)
		//     return
		// }

		// Placeholder response
		response := &models.BillingHistoryResponse{
			Records: []models.BillingRecord{},
			Total:   0,
			Limit:   req.Limit,
			Offset:  req.Offset,
		}

		writeJSONResponse(w, http.StatusOK, response)
	}
}

// CalculatePricing handles pricing calculation requests
func CalculatePricing(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement pricing calculation endpoint
		// This would use the pricing engine to calculate costs for given requirements

		response := map[string]interface{}{
			"message": "Pricing calculation endpoint not yet implemented",
			"status":  "coming_soon",
		}

		writeJSONResponse(w, http.StatusNotImplemented, response)
	}
}

// GetPricingRates handles pricing rates requests
func GetPricingRates(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement pricing rates endpoint
		// This would return current pricing rates for different GPU models

		response := map[string]interface{}{
			"message": "Pricing rates endpoint not yet implemented",
			"status":  "coming_soon",
		}

		writeJSONResponse(w, http.StatusNotImplemented, response)
	}
}

// GetProviderEarnings handles provider earnings requests
func GetProviderEarnings(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerIDStr := chi.URLParam(r, "providerID")
		providerID, err := uuid.Parse(providerIDStr)
		if err != nil {
			logger.Error("Invalid provider ID", zap.String("provider_id", providerIDStr), zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid provider ID", err)
			return
		}

		// TODO: Implement provider earnings retrieval
		// earnings, err := billingService.GetProviderEarnings(r.Context(), providerID)
		// if err != nil {
		//     logger.Error("Failed to get provider earnings", zap.String("provider_id", providerIDStr), zap.Error(err))
		//     writeErrorResponse(w, http.StatusInternalServerError, "Failed to get provider earnings", err)
		//     return
		// }

		// Placeholder response
		response := &models.ProviderEarningsResponse{
			ProviderID:      providerID,
			TotalEarnings:   decimal.Zero,
			PendingEarnings: decimal.Zero,
			PaidEarnings:    decimal.Zero,
			TotalSessions:   0,
			TotalHours:      decimal.Zero,
			AvgHourlyRate:   decimal.Zero,
			Period:          "all_time",
		}

		writeJSONResponse(w, http.StatusOK, response)
	}
}

// RequestPayout handles provider payout requests
func RequestPayout(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerIDStr := chi.URLParam(r, "providerID")
		providerID, err := uuid.Parse(providerIDStr)
		if err != nil {
			logger.Error("Invalid provider ID", zap.String("provider_id", providerIDStr), zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid provider ID", err)
			return
		}

		var req models.PayoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode payout request", zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
			return
		}

		req.ProviderWalletID = providerID // This should be the wallet ID, not provider ID

		// TODO: Implement payout processing
		// payout, err := billingService.ProcessPayout(r.Context(), &req)
		// if err != nil {
		//     logger.Error("Failed to process payout", zap.Error(err))
		//     writeErrorResponse(w, http.StatusInternalServerError, "Failed to process payout", err)
		//     return
		// }

		response := map[string]interface{}{
			"message":     "Payout request received",
			"provider_id": providerID,
			"amount":      req.Amount,
			"to_address":  req.ToAddress,
			"status":      "pending",
		}

		logger.Info("Payout requested",
			zap.String("provider_id", providerIDStr),
			zap.String("amount", req.Amount.String()),
			zap.String("to_address", req.ToAddress),
		)

		writeJSONResponse(w, http.StatusAccepted, response)
	}
}

// GetProviderRates handles provider rates requests
func GetProviderRates(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerIDStr := chi.URLParam(r, "providerID")
		providerID, err := uuid.Parse(providerIDStr)
		if err != nil {
			logger.Error("Invalid provider ID", zap.String("provider_id", providerIDStr), zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid provider ID", err)
			return
		}

		// TODO: Implement provider rates retrieval
		response := map[string]interface{}{
			"message":     "Provider rates endpoint not yet implemented",
			"provider_id": providerID,
			"status":      "coming_soon",
		}

		writeJSONResponse(w, http.StatusNotImplemented, response)
	}
}

// SetProviderRates handles provider rates update requests
func SetProviderRates(billingService *service.BillingService, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerIDStr := chi.URLParam(r, "providerID")
		providerID, err := uuid.Parse(providerIDStr)
		if err != nil {
			logger.Error("Invalid provider ID", zap.String("provider_id", providerIDStr), zap.Error(err))
			writeErrorResponse(w, http.StatusBadRequest, "Invalid provider ID", err)
			return
		}

		// TODO: Implement provider rates update
		response := map[string]interface{}{
			"message":     "Provider rates update endpoint not yet implemented",
			"provider_id": providerID,
			"status":      "coming_soon",
		}

		writeJSONResponse(w, http.StatusNotImplemented, response)
	}
}
