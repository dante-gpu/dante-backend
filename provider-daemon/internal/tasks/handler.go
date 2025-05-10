package tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dante-gpu/dante-backend/provider-daemon/internal/config"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/executor"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/models"

	// "github.com/dante-gpu/dante-backend/provider-daemon/internal/reporting" // Not used yet
	"go.uber.org/zap"
)

// NatsStatusPublisher defines an interface for publishing status updates.
// This helps decouple the task handler from the concrete NATS client implementation.
type NatsStatusPublisher interface {
	PublishStatus(statusUpdate *models.TaskStatusUpdate) error
}

// Handler orchestrates task execution and status reporting.
type Handler struct {
	logger   *zap.Logger
	cfg      *config.Config
	reporter NatsStatusPublisher // Using the interface for reporting
	executor executor.Executor   // Interface for task execution
}

// NewHandler creates a new task handler.
func NewHandler(cfg *config.Config, logger *zap.Logger, statusPublisher NatsStatusPublisher, exec executor.Executor) *Handler {
	return &Handler{
		logger:   logger,
		cfg:      cfg,
		reporter: statusPublisher,
		executor: exec,
	}
}

// SetReporter allows setting the NatsStatusPublisher after Handler initialization.
// This is useful to break initialization cycles.
func (h *Handler) SetReporter(reporter NatsStatusPublisher) {
	h.reporter = reporter
}

// HandleTask is the entry point for processing a received task.
// It creates a workspace, executes the task using the configured executor,
// and reports status updates (InProgress, Completed, Failed).
func (h *Handler) HandleTask(task *models.Task) error {
	h.logger.Info("Handling task", zap.String("job_id", task.JobID), zap.String("job_type", task.JobType))

	// 1. Create a unique workspace for the task
	// Workspace path: {cfg.WorkspaceDir}/{job_id}
	workspacePath := filepath.Join(h.cfg.WorkspaceDir, task.JobID)
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		h.logger.Error("Failed to create task workspace", zap.String("job_id", task.JobID), zap.String("workspace", workspacePath), zap.Error(err))
		// Report failure status immediately if workspace creation fails
		_ = h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusFailed, fmt.Sprintf("Failed to create workspace: %v", err)))
		return fmt.Errorf("failed to create workspace for job %s: %w", task.JobID, err)
	}
	h.logger.Info("Task workspace created", zap.String("job_id", task.JobID), zap.String("path", workspacePath))

	// Defer cleanup of the workspace
	defer func() {
		if err := os.RemoveAll(workspacePath); err != nil {
			h.logger.Error("Failed to cleanup task workspace", zap.String("job_id", task.JobID), zap.String("workspace", workspacePath), zap.Error(err))
		} else {
			h.logger.Info("Task workspace cleaned up", zap.String("job_id", task.JobID), zap.String("path", workspacePath))
		}
	}()

	// 2. Report InProgress status
	statusInProgress := models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusInProgress, "Task execution started.")
	if err := h.reporter.PublishStatus(statusInProgress); err != nil {
		h.logger.Warn("Failed to publish InProgress status update", zap.String("job_id", task.JobID), zap.Error(err))
		// Continue with execution even if status update fails for now
	}

	// 3. Execute the task using the executor
	// Context for execution (can be enhanced with task-specific timeouts if needed)
	execCtx, cancelExec := context.WithCancel(context.Background()) // Basic context
	// If task.JobParams has a timeout, it will be handled by the ScriptExecutor itself.
	// This context is more for cancelling the overall HandleTask flow if needed, or could pass down to executor.
	defer cancelExec()

	executionResult := h.executor.Execute(execCtx, task, workspacePath, h.logger)

	// 4. Report final status based on execution result
	var finalStatus *models.TaskStatusUpdate
	stdoutSnippet := getSnippet(executionResult.Stdout, 512) // Max 512 chars for snippet
	stderrSnippet := getSnippet(executionResult.Stderr, 512)

	if executionResult.ExitCode == 0 && executionResult.Error == nil {
		finalStatus = models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusCompleted, "Task completed successfully.")
		finalStatus.ExitCode = &executionResult.ExitCode
		finalStatus.ExecutionLog = fmt.Sprintf("STDOUT:\n%s\n\nSTDERR:\n%s", stdoutSnippet, stderrSnippet)
		h.logger.Info("Task completed successfully", zap.String("job_id", task.JobID), zap.Int("exit_code", executionResult.ExitCode))
	} else {
		errMsg := fmt.Sprintf("Task failed. Exit Code: %d.", executionResult.ExitCode)
		if executionResult.Error != nil {
			errMsg = fmt.Sprintf("Task failed: %v. Exit Code: %d.", executionResult.Error, executionResult.ExitCode)
		}
		finalStatus = models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusFailed, errMsg)
		finalStatus.ExitCode = &executionResult.ExitCode
		finalStatus.ExecutionLog = fmt.Sprintf("STDOUT:\n%s\n\nSTDERR:\n%s", stdoutSnippet, stderrSnippet)
		h.logger.Error("Task failed", zap.String("job_id", task.JobID), zap.Int("exit_code", executionResult.ExitCode), zap.Error(executionResult.Error))
	}

	if err := h.reporter.PublishStatus(finalStatus); err != nil {
		h.logger.Error("Failed to publish final status update", zap.String("job_id", task.JobID), zap.Error(err))
		// This is a critical failure if the final status cannot be reported.
		// However, the task itself has finished. The error here is about reporting.
		return fmt.Errorf("task execution finished but failed to report final status for job %s: %w", task.JobID, err)
	}

	// If the task itself failed, return that error to the NATS handler so it can NAK if needed
	if executionResult.Error != nil {
		return fmt.Errorf("task execution failed for job %s: %w", task.JobID, executionResult.Error)
	}

	return nil
}

// getSnippet returns a snippet of a string, up to a max length.
func getSnippet(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "... (truncated)"
}
