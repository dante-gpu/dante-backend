package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dante-gpu/dante-backend/api-gateway/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// I need to set up the router.
	r := chi.NewRouter()

	// I should add basic middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// Replace chi logger with my custom Zap logger middleware
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.RequestTimeout))

	// I'll define a simple health check endpoint for now.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Use Zap logger for endpoint logging
		logger.Info("Health check endpoint hit", zap.String("path", r.URL.Path))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "API Gateway is healthy")
	})

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
