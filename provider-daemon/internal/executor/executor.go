package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dante-gpu/dante-backend/provider-daemon/internal/models"
	"go.uber.org/zap"
)

// ExecutionResult holds the outcome of a task execution.
type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error // For errors during the execution setup or process itself, not script errors
}

// Executor defines the interface for running tasks.
type Executor interface {
	Execute(ctx context.Context, task *models.Task, workspacePath string, logger *zap.Logger) ExecutionResult
}

// ScriptExecutor implements the Executor interface for running shell scripts.
type ScriptExecutor struct{}

// NewScriptExecutor creates a new ScriptExecutor.
func NewScriptExecutor() *ScriptExecutor {
	return &ScriptExecutor{}
}

// Execute runs a script defined in the task's parameters.
// It expects task.JobParams to contain:
// - "script_content": string (the script)
// - "script_interpreter": string (e.g., "/bin/bash", "python3")
// - "script_filename": string (e.g., "run.sh", "main.py" - optional, defaults if not provided)
// - "timeout_seconds": int (optional, task execution timeout)
func (se *ScriptExecutor) Execute(ctx context.Context, task *models.Task, workspacePath string, logger *zap.Logger) ExecutionResult {
	logger.Info("Starting script execution", zap.String("job_id", task.JobID), zap.String("workspace", workspacePath))

	scriptContent, ok := task.JobParams["script_content"].(string)
	if !ok || strings.TrimSpace(scriptContent) == "" {
		logger.Error("Script content not found or empty in task parameters", zap.String("job_id", task.JobID))
		return ExecutionResult{Error: fmt.Errorf("script_content is required and cannot be empty"), ExitCode: -1}
	}

	interpreter, ok := task.JobParams["script_interpreter"].(string)
	if !ok || strings.TrimSpace(interpreter) == "" {
		logger.Warn("Script interpreter not specified, defaulting to /bin/sh", zap.String("job_id", task.JobID))
		interpreter = "/bin/sh" // Default interpreter
	}

	scriptFilename, ok := task.JobParams["script_filename"].(string)
	if !ok || strings.TrimSpace(scriptFilename) == "" {
		if interpreter == "python" || interpreter == "python3" {
			scriptFilename = "main.py"
		} else {
			scriptFilename = "task_script.sh"
		}
	}
	scriptPath := filepath.Join(workspacePath, scriptFilename)

	// Write the script content to a file in the workspace
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755) // Make it executable
	if err != nil {
		logger.Error("Failed to write script to workspace", zap.String("job_id", task.JobID), zap.Error(err))
		return ExecutionResult{Error: fmt.Errorf("failed to write script file: %w", err), ExitCode: -1}
	}
	logger.Info("Script written to file", zap.String("path", scriptPath))

	var execCtx context.Context
	var cancel context.CancelFunc

	timeoutSeconds, hasTimeout := task.JobParams["timeout_seconds"].(float64) // YAML might parse numbers as float64
	if hasTimeout && timeoutSeconds > 0 {
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		logger.Info("Script execution timeout set", zap.Float64("seconds", timeoutSeconds))
	} else {
		execCtx, cancel = context.WithCancel(ctx) // No timeout or indefinite
	}
	defer cancel()

	cmd := exec.CommandContext(execCtx, interpreter, scriptPath)
	cmd.Dir = workspacePath // Execute from the workspace directory

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	logger.Info("Executing script", zap.String("interpreter", interpreter), zap.String("script", scriptPath))

	runErr := cmd.Run()
	duration := time.Since(startTime)
	logger.Info("Script execution finished", zap.Duration("duration", duration), zap.Error(runErr))

	result := ExecutionResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Errorf("script exited with code %d: %w", result.ExitCode, exitErr) // Capture ExitError
		} else if execCtx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("script execution timed out after %v", time.Duration(timeoutSeconds)*time.Second)
			result.ExitCode = -2 // Specific code for timeout
			logger.Warn("Script execution timed out", zap.String("job_id", task.JobID))
		} else {
			result.Error = fmt.Errorf("script execution failed: %w", runErr)
			result.ExitCode = -1 // Generic error
			logger.Error("Script execution failed with non-ExitError", zap.String("job_id", task.JobID), zap.Error(runErr))
		}
	} else {
		result.ExitCode = 0
	}

	logger.Debug("Script execution details",
		zap.Int("exit_code", result.ExitCode),
		zap.String("stdout_len", fmt.Sprintf("%d bytes", len(result.Stdout))),
		zap.String("stderr_len", fmt.Sprintf("%d bytes", len(result.Stderr))),
	)

	return result
}

// TODO: Implement DockerExecutor
// type DockerExecutor struct {
//    Client *client.Client // Docker client
//    Logger *zap.Logger
// }
// func (de *DockerExecutor) Execute(...) ExecutionResult { ... }
