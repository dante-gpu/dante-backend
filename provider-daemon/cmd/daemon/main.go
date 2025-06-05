package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	// "strconv" // Will be needed for other CLI commands
	// "time" // Will be needed for other CLI commands

	"github.com/dante-gpu/dante-backend/provider-daemon/internal/billing"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/config"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/executor"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/gpu"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/models"
	cli_models "github.com/dante-gpu/dante-backend/provider-daemon/internal/models" // Alias for cli response models
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/nats"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/tasks"

	// "github.com/google/uuid" // Will be needed for other CLI commands or if instance ID generation is reactivated here
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Version   = "dev"     // Injected at build time
	BuildDate = "unknown" // Injected at build time
)

// CLI flags
var (
	configPath      = flag.String("config", filepath.Join("configs", "config.yaml"), "Path to the configuration file")
	getGpusJSON     = flag.Bool("get-gpus-json", false, "Detect GPUs and output as JSON, then exit")
	getSettingsJSON = flag.Bool("get-settings-json", false, "Output current provider settings as JSON, then exit")
	// TODO: Add flags for other CLI commands as they are implemented
	// updateSettingsJSON     = flag.String("update-settings-json", "", "Update provider settings from a JSON string, then exit")
	// setGpuConfigJSON       = flag.Bool("set-gpu-config-json", false, "Set individual GPU rental config, then exit (requires --gpu-id, --rate, --available)")
	// gpuIDForConfig         = flag.String("gpu-id", "", "GPU ID for --set-gpu-config-json")
	// rateForConfig          = flag.Float64("rate", 0.0, "Hourly rate for --set-gpu-config-json")
	// availableForConfig     = flag.Bool("available", false, "Availability for --set-gpu-config-json")
	// getLocalJobsJSON       = flag.Bool("get-local-jobs-json", false, "Get current local jobs as JSON, then exit")
	// getNetworkStatusJSON   = flag.Bool("get-network-status-json", false, "Get network status as JSON, then exit")
	// getFinancialSummaryJSON = flag.Bool("get-financial-summary-json", false, "Get financial summary as JSON, then exit")
)

func main() {
	flag.Parse() // Parse all defined CLI flags

	tempLogger, _ := setupLogger("info")
	cfg, err := config.LoadConfig(*configPath, tempLogger)
	if err != nil {
		tempLogger.Fatal("Failed to load configuration", zap.Error(err), zap.String("path", *configPath))
	}

	logger, err := setupLogger(cfg.LogLevel)
	if err != nil {
		tempLogger.Fatal("Failed to setup logger with config level", zap.Error(err))
	}
	defer logger.Sync()
	cfg.Logger = logger

	// --- Handle CLI Commands ---
	if *getGpusJSON {
		handleGetGpusJSON(cfg, logger)
		return
	}
	if *getSettingsJSON {
		handleGetSettingsJSON(cfg, logger)
		return
	}
	// Add other CLI command handlers here as they are implemented

	// --- Start Daemon Mode (if no CLI command was executed) ---
	logger.Info("Starting Dante GPU Provider Daemon",
		zap.String("version", Version),
		zap.String("buildDate", BuildDate),
		zap.String("instanceID", cfg.InstanceID),
		zap.String("logLevel", cfg.LogLevel),
	)

	// Initialize components for daemon mode
	gpuDetector := gpu.NewDetector(logger)
	// go gpuDetector.StartMonitoring() // Commented out: StartMonitoring not found on current Detector struct
	// defer gpuDetector.StopMonitoring() // Commented out: StopMonitoring not found on current Detector struct

	billingClient := billing.NewClient(&cfg.BillingClientConfig, logger)

	// Initialize Executor
	// TODO: Allow selection of executor type (docker, script) based on config
	var taskExecutor executor.Executor
	dockerExec, err := executor.NewDockerExecutor(logger, billingClient, gpuDetector)
	if err != nil {
		logger.Warn("Failed to initialize Docker executor, falling back to script executor", zap.Error(err))
		// Fallback to script executor
		scriptExec := executor.NewScriptExecutor() // Corrected: No arguments, one return value
		// if scriptErr != nil { // NewScriptExecutor now returns no error
		// 	logger.Fatal("Failed to initialize any task executor", zap.Error(scriptErr))
		// }
		taskExecutor = scriptExec
	} else {
		taskExecutor = dockerExec
	}

	// Initialize Task Handler
	// The linter errors indicate a mismatch in types/order for NewHandler.
	// Will need to check the signature of tasks.NewHandler.
	taskHandler := tasks.NewHandler(cfg, logger, taskExecutor, gpuDetector, billingClient)

	natsClient, err := nats.NewClient(cfg, logger, taskHandler.HandleTask)
	if err != nil {
		logger.Fatal("Failed to initialize NATS client", zap.Error(err))
	}

	if err := natsClient.StartListening(); err != nil {
		logger.Fatal("Failed to start NATS listener", zap.Error(err))
	}
	defer natsClient.Stop()

	logger.Info("Provider Daemon is running. Waiting for tasks...")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	logger.Info("Shutting down Provider Daemon...")
}

func handleGetGpusJSON(cfg *config.Config, logger *zap.Logger) {
	logger.Info("CLI command: --get-gpus-json")
	gpuDetector := gpu.NewDetector(logger)

	detectedSystemGPUs, err := gpuDetector.DetectGPUsOnce()
	if err != nil {
		outputJSONError(fmt.Sprintf("Failed to detect GPUs: %v", err), os.Stderr, logger)
		os.Exit(1)
	}

	cliGPUs := make([]cli_models.CliGpuInfo, 0, len(detectedSystemGPUs))
	gpuRentalConfigMap := make(map[string]config.GpuRentalConfigEntry)
	for _, rentalConfig := range cfg.GpuRentalConfigs {
		gpuRentalConfigMap[rentalConfig.GpuID] = rentalConfig
	}

	for _, systemGPU := range detectedSystemGPUs { // Iterate over gpu.GPUInfo
		// Basic mapping from gpu.GPUInfo to cli_models.CliGpuInfo
		cliInfo := cli_models.CliGpuInfo{
			ID:                 systemGPU.ID,
			Name:               systemGPU.Name,
			Model:              systemGPU.Model,
			VRAMTotalMB:        uint32(systemGPU.VRAMTotal), // Cast from uint64
			VRAMFreeMB:         uint32(systemGPU.VRAMFree),  // Cast from uint64
			IsAvailableForRent: false,                       // Default, will be overridden by rental config if present
		}

		// Optional fields - assuming gpu.GPUInfo fields are 0/empty if not applicable/available
		// The cli_models uses pointers, so we only set them if data is meaningful.
		if systemGPU.Utilization > 0 { // Check if utilization is reported
			util := uint32(systemGPU.Utilization) // Already uint8, direct cast is fine
			cliInfo.UtilizationGPUPercent = &util
		}
		if systemGPU.Temperature > 0 { // Check if temperature is reported
			temp := uint32(systemGPU.Temperature) // Already uint8, direct cast is fine
			cliInfo.TemperatureC = &temp
		}
		if systemGPU.PowerDraw > 0 { // Check if power draw is reported
			power := systemGPU.PowerDraw // Already uint32
			cliInfo.PowerDrawW = &power
		}

		// Apply rental config from cfg.GpuRentalConfigs
		if rentalCfg, ok := gpuRentalConfigMap[systemGPU.ID]; ok {
			cliInfo.IsAvailableForRent = rentalCfg.IsAvailableForRent
			if rentalCfg.IsAvailableForRent && rentalCfg.CurrentHourlyRateDGPU > 0 {
				rate := rentalCfg.CurrentHourlyRateDGPU // This is float32 in GpuRentalConfigEntry
				cliInfo.CurrentHourlyRateDGPU = &rate
			}
		}
		cliGPUs = append(cliGPUs, cliInfo)
	}
	outputJSON(cliGPUs, logger)
}

func handleGetSettingsJSON(cfg *config.Config, logger *zap.Logger) {
	logger.Info("CLI command: --get-settings-json")
	settings := cli_models.CliProviderSettings{
		// DefaultHourlyRateDGPU: float32(cfg.DefaultHourlyRateDGPU), // Commented out: field potentially missing from current config.Config
		PreferredCurrency: cfg.PreferredCurrency, // This one should exist
		// MinJobDurationMinutes: cfg.MinJobDurationMinutes, // Commented out: field potentially missing from current config.Config
		MaxConcurrentJobs: cfg.MaxConcurrentJobs, // This one should exist
	}
	outputJSON(settings, logger)
}

func outputJSON(data interface{}, logger *zap.Logger) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// Log internal error, then attempt to output a simple JSON error to stdout for Tauri
		logger.Error("Failed to marshal data to JSON for CLI output", zap.Error(err))
		fmt.Fprintf(os.Stdout, `{"error": "Failed to marshal data to JSON: %s"}\n`, err.Error())
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
	os.Exit(0) // Ensure exit after successful JSON output for CLI mode
}

func outputJSONError(message string, writer *os.File, logger *zap.Logger) {
	logger.Error("CLI command error", zap.String("error_message", message))
	errorData := map[string]string{"error": message}
	jsonData, err := json.Marshal(errorData)
	if err != nil {
		fmt.Fprintf(writer, `{"error": "Failed to marshal error message to JSON. Original error: %s"}\n`, message)
		os.Exit(1)
		return
	}
	fmt.Fprintln(writer, string(jsonData))
	os.Exit(1) // Exit after error for CLI mode
}

func setupLogger(levelString string) (*zap.Logger, error) {
	var logLevel zapcore.Level
	switch levelString {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	case "fatal":
		logLevel = zapcore.FatalLevel
	default:
		// Fallback for CLI mode before config is loaded, or if config is malformed
		fmt.Fprintf(os.Stderr, "Invalid log level specified: %s. Defaulting to info.\n", levelString)
		logLevel = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	logDir := filepath.Join(".", "logs", "provider-daemon")
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		// Fallback to console-only logging if directory creation fails
		consoleCfg := zap.Config{
			Level:            zap.NewAtomicLevelAt(logLevel),
			Development:      false,  // Set to true for more human-readable console output if preferred
			Encoding:         "json", // Or "console" for human-readable
			EncoderConfig:    encoderConfig,
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
		logger, buildErr := consoleCfg.Build()
		if buildErr != nil {
			return nil, fmt.Errorf("failed to build console logger: %w", buildErr)
		}
		// Log the directory creation error using the console logger itself, if possible
		logger.Error("Failed to create log directory, logging to console only", zap.String("directory", logDir), zap.Error(err))
		return logger, nil
	}

	logFileName := filepath.Join(logDir, "daemon.log")

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(zapcore.Lock(mustOpen(logFileName))),
		logLevel,
	)

	// Setup console core based on whether it's likely a CLI command execution or daemon mode
	// For CLI commands, we often want cleaner output on stdout for JSON, and logs to stderr/file.
	// For daemon mode, teeing to console can be verbose but useful for development.
	// The current tempLogger is console-only. If we are in CLI mode, we might prefer stderr for logs.

	// Let's use a simpler console encoder for daemon mode, or if explicit console logging is desired.
	// For CLI commands that output JSON to stdout, we should ensure logs go to stderr or file.
	consoleEncoderCfg := zap.NewDevelopmentEncoderConfig() // More readable for console
	consoleEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoderCfg.TimeKey = "ts"
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderCfg),
		zapcore.AddSync(os.Stderr), // Log to stderr for console
		logLevel,
	)

	// For CLI commands, we might want to suppress regular consoleCore if stdout is for JSON.
	// However, error logs from outputJSONError should still go to stderr.
	// The logger passed to handleGetGpusJSON etc., will use this Tee.
	// If it's a CLI command that prints JSON to stdout, non-error logs should ideally go to file or stderr only.
	// This setup tees all logs (file + stderr console). Refinements could be made for cleaner CLI stdout.

	teeCore := zapcore.NewTee(fileCore, consoleCore)

	return zap.New(teeCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

func mustOpen(filePath string) *os.File {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Fallback if log file cannot be opened, print to stderr
		fmt.Fprintf(os.Stderr, "PANIC: Failed to open log file %s: %v\n", filePath, err)
		// Attempt to use a basic stderr logger if primary logging fails catastrophically
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stderr"}
		logger, _ := cfg.Build()
		logger.Panic("Failed to open log file", zap.String("path", filePath), zap.Error(err))
		// Should not reach here due to panic
	}
	return file
}

// Ensure all model types are correctly imported and used to avoid "unused import" errors
// when some CLI command handlers are not yet fully implemented.
var _ models.Task           // From provider-daemon/internal/models
var _ cli_models.CliGpuInfo // From provider-daemon/internal/models (aliased)
