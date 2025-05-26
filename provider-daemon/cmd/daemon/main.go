package main

import (
	"fmt" // Added for logger setup error formatting
	// Standard log for initial errors
	"os"
	"os/signal"
	"syscall"

	"github.com/dante-gpu/dante-backend/provider-daemon/internal/config"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/executor" // Added executor import
	nats_client "github.com/dante-gpu/dante-backend/provider-daemon/internal/nats"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/tasks"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Load Configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize Logger
	logger, err := setupLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() // Flush buffer

	logger.Info("Starting Provider Daemon", zap.String("instance_id", cfg.InstanceID))

	// Create a new ScriptExecutor
	scriptExec := executor.NewScriptExecutor()

	// Create a new DockerExecutor
	dockerExec, err := executor.NewDockerExecutor(logger)
	if err != nil {
		// Log the error from NewDockerExecutor and decide if it's fatal.
		// If Docker is essential, os.Exit(1) might be appropriate.
		// For now, log it and the handler will fail tasks requiring Docker.
		logger.Error("Failed to initialize Docker executor. Docker tasks will fail.", zap.Error(err))
		// We can proceed without a dockerExec if script tasks are still valuable.
		// If dockerExec is nil, the handler will report an error if a Docker task is received.
	}

	// Create Task Handler, passing both executors.
	// The NATS client (for status reporting) will be set later to break init cycle.
	taskHandlerInstance := tasks.NewHandler(cfg, logger, nil, scriptExec, dockerExec)

	// Initialize NATS Client
	// The NATS client needs the task handler to process incoming messages.
	natsClient, err := nats_client.NewClient(cfg, logger, taskHandlerInstance.HandleTask)
	if err != nil {
		logger.Fatal("Failed to initialize NATS client", zap.Error(err))
	}

	// Set the NATS client as the status reporter for the task handler
	// This completes the dependency cycle: NATS client uses TaskHandler, TaskHandler uses NATS client (for reporting).
	taskHandlerInstance.SetReporter(natsClient)

	// Start NATS listener
	if err := natsClient.StartListening(); err != nil {
		logger.Fatal("Failed to start NATS listener", zap.Error(err))
	}

	logger.Info("Provider Daemon initialized and NATS listener started. Waiting for tasks...")

	// Graceful Shutdown Handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received

	logger.Info("Shutdown signal received, starting graceful shutdown...")

	// Stop NATS client
	natsClient.Stop()

	// Perform other cleanup if necessary
	logger.Info("Provider Daemon shut down gracefully.")
}

// setupLogger configures Zap based on the log level string.
func setupLogger(levelString string) (*zap.Logger, error) {
	var logLevel zapcore.Level
	if err := logLevel.Set(levelString); err != nil {
		logLevel = zapcore.InfoLevel // Default to info if parsing fails
	}

	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false, // Set to true for more dev-friendly output
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

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err) // Ensure fmt is imported
	}

	return logger, nil
}
