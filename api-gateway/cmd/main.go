package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dante-gpu/dante-backend/api-gateway/internal/config"
	"github.com/dante-gpu/dante-backend/api-gateway/internal/handlers"
	customMiddleware "github.com/dante-gpu/dante-backend/api-gateway/internal/middleware" // Alias to avoid conflict
	nats_client "github.com/dante-gpu/dante-backend/api-gateway/internal/nats"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go" // Import nats package
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// I should load the configuration first.
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		// Using a temporary basic logger for config loading errors
		basicLogger, _ := zap.NewProduction()
		basicLogger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// I need to set up the Zap logger based on config.
	logger, err := setupLogger(cfg.LogLevel)
	if err != nil {
		// Use standard log if Zap setup fails initially
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync() // Flushing buffer, important!

	// == Establish NATS Connection ==
	nc, err := nats_client.Connect(cfg.NatsAddress, logger)
	if err != nil {
		// Log the error but maybe allow the server to start without NATS?
		// For now, let's make it fatal as job submission is core.
		logger.Fatal("Failed to establish initial NATS connection", zap.Error(err))
	}
	defer nc.Close() // Ensure NATS connection is closed on shutdown

	// Optional: Connect to JetStream if needed
	// js, err := nats_client.ConnectJetStream(nc, logger)
	// if err != nil {
	// 	logger.Fatal("Failed to establish NATS JetStream connection", zap.Error(err))
	// }

	// I need to set up the router.
	r := chi.NewRouter()

	// I should add basic middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// Replace chi logger with my custom Zap logger middleware
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.RequestTimeout))

	// I need to create instances of my handlers.
	authHandler := handlers.NewAuthHandler(logger, cfg)
	jobHandler := handlers.NewJobHandler(logger, cfg, nc) // Pass NATS connection

	// == Public Routes ==
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check NATS connection status as part of health <3 ?
		healthStatus := http.StatusOK
		healthMsg := "API Gateway is healthy"
		if nc.Status() != nats.CONNECTED {
			healthStatus = http.StatusServiceUnavailable
			healthMsg = "API Gateway is running, but NATS connection is down"
			logger.Warn("Health check reporting NATS is not connected", zap.String("nats_status", nc.Status().String()))
		}

		logger.Info("Health check endpoint hit",
			zap.String("path", r.URL.Path),
			zap.String("nats_status", nc.Status().String()),
		)
		w.WriteHeader(healthStatus)
		fmt.Fprintf(w, healthMsg)
	})

	// == Authentication Routes ==
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)

		// Routes requiring authentication
		r.Group(func(r chi.Router) {
			// I need to apply the JWT authentication middleware here.
			r.Use(customMiddleware.Authenticator(logger, cfg.JwtSecret))
			r.Get("/profile", authHandler.Profile)
		})
	})

	// == API V1 Routes (Protected) ==
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware to all v1 routes
		r.Use(customMiddleware.Authenticator(logger, cfg.JwtSecret))

		// Job submission routes
		r.Post("/jobs", jobHandler.SubmitJob)
		r.Get("/jobs/{jobID}", jobHandler.GetJobStatus)
		r.Delete("/jobs/{jobID}", jobHandler.CancelJob)

		// Admin routes (placeholder)
		// r.Group(func(r chi.Router) {
		//    r.Use(customMiddleware.RequireRole("admin")) // Role checking middleware needed
		//    r.Get("/admin-stats", handlers.GetAdminStats)
		// })
	})

	// == Service Proxy Route (Example - potentially protected) ==
	// proxyHandler := handlers.NewProxyHandler(logger, cfg) // Needs Consul client etc.
	// r.HandleFunc("/services/{serviceName}/*", proxyHandler.ServeHTTP)

	// == Admin Dashboard Route (Example - protected) ==
	// r.Group(func(r chi.Router) {
	//     r.Use(customMiddleware.Authenticator(logger, cfg.JwtSecret))
	//     r.Use(customMiddleware.RequireRole("admin")) // Role checking middleware needed
	//     r.Get("/admin", handlers.AdminDashboard)
	// })

	// I need to start the HTTP server.
	logger.Info("Starting API Gateway", zap.String("port", cfg.Port))
	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// setupLogger configures Zap based on the log level string.
func setupLogger(levelString string) (*zap.Logger, error) {
	var logLevel zapcore.Level
	if err := logLevel.Set(levelString); err != nil {
		logLevel = zapcore.InfoLevel // Default to info if parsing fails
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Encoding:    "json", // Or "console" for more readable output during dev
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
		OutputPaths:      []string{"stdout"}, // Log to standard output
		ErrorOutputPaths: []string{"stderr"}, // Log errors to standard error
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// NewStructuredLogger returns a middleware that logs request details using Zap.
func NewStructuredLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor) // To capture status code

			defer func() {
				duration := time.Since(start)
				// Log details for the request
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
