package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Use imports relative to this service's module path
	"github.com/dante-gpu/dante-backend/provider-registry-service/internal/config"
	consul_client "github.com/dante-gpu/dante-backend/provider-registry-service/internal/consul"
	"github.com/dante-gpu/dante-backend/provider-registry-service/internal/handlers"
	"github.com/dante-gpu/dante-backend/provider-registry-service/internal/server"
	"github.com/dante-gpu/dante-backend/provider-registry-service/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// --- Configuration ---
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err) // Use standard log before Zap is up
	}

	// --- Logger ---
	logger, err := setupLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync() // Flush logs before exiting

	// --- Consul Client ---
	consulClient, err := consul_client.Connect(cfg.ConsulAddress, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Consul agent", zap.Error(err))
	}

	// --- Consul Service Registration ---
	// Generate a unique ID for this service instance
	serviceID := config.GenerateServiceID(cfg.ServiceIDPrefix)
	logger.Info("Generated unique service ID for Consul", zap.String("service_id", serviceID))

	err = consul_client.RegisterService(consulClient, cfg, serviceID, logger)
	if err != nil {
		logger.Fatal("Failed to register service with Consul", zap.Error(err))
	}
	logger.Info("Successfully registered service with Consul",
		zap.String("service_name", cfg.ServiceName),
		zap.String("service_id", serviceID),
	)

	// --- Initialize Store ---
	providerStore := store.NewInMemoryProviderStore()
	logger.Info("In-memory provider store initialized")

	// --- Setup Router and Server ---
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// Add Zap logging middleware (using a similar one as api-gateway)
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.RequestTimeout))

	// Add Health Check endpoint (required by Consul registration)
	r.Get(cfg.HealthCheckPath, func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add more detailed health checks (e.g., DB connectivity)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Provider Registry Service is healthy")
		logger.Debug("Health check endpoint hit")
	})

	// --- Mount API Handlers ---
	providerHandler := handlers.NewProviderHandler(logger, cfg, providerStore)
	r.Mount("/providers", providerHandler.Routes())
	logger.Info("Provider API routes mounted under /providers")

	srv := server.NewServer(cfg.Port, r, logger)

	// --- Start Server Goroutine ---
	go func() {
		logger.Info("Starting Provider Registry Service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Could not listen on port", zap.String("port", cfg.Port), zap.Error(err))
		}
	}()

	// --- Graceful Shutdown & Consul Deregistration ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received
	logger.Info("Shutdown signal received, starting graceful shutdown...")

	// Deregister from Consul
	logger.Info("Deregistering service from Consul", zap.String("service_id", serviceID))
	if err := consulClient.Agent().ServiceDeregister(serviceID); err != nil {
		logger.Error("Failed to deregister service from Consul", zap.String("service_id", serviceID), zap.Error(err))
	} else {
		logger.Info("Successfully deregistered service from Consul", zap.String("service_id", serviceID))
	}

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown uncleanly", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}

// setupLogger configures Zap based on the log level string.
// (Identical to the one in api-gateway, maybe move to a shared lib later?)
func setupLogger(levelString string) (*zap.Logger, error) {
	var logLevel zapcore.Level
	if err := logLevel.Set(levelString); err != nil {
		logLevel = zapcore.InfoLevel // Default to info if parsing fails
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// NewStructuredLogger returns a middleware that logs request details using Zap.
// (Identical to the one in api-gateway, maybe move to a shared lib later?)
func NewStructuredLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start)
				logger.Info("Request completed",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_ip", r.RemoteAddr),
					zap.String("request_id", middleware.GetReqID(r.Context())),
					zap.Int("status", ww.Status()),
					zap.Int("bytes", ww.BytesWritten()),
					zap.Duration("duration", duration),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
