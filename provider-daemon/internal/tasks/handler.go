package tasks

import (
	"context"
	"fmt"
	"os"

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
	logger         *zap.Logger
	cfg            *config.Config
	reporter       NatsStatusPublisher // Using the interface for reporting
	scriptExecutor executor.Executor   // Specifically for script tasks
	dockerExecutor executor.Executor   // Specifically for Docker tasks
}

// NewHandler creates a new task handler.
// It requires a logger, config, a status publisher (like the NATS client),
// and instances of script and Docker executors.
func NewHandler(cfg *config.Config, logger *zap.Logger, statusPublisher NatsStatusPublisher, scriptExec executor.Executor, dockerExec executor.Executor) *Handler {
	return &Handler{
		logger:         logger,
		cfg:            cfg,
		reporter:       statusPublisher,
		scriptExecutor: scriptExec,
		dockerExecutor: dockerExec, // Store the Docker executor
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
	ctx := context.Background() // Create a new context for this task handling
	// Log the received task with its execution type
	h.logger.Info("Received task for handling",
		zap.String("job_id", task.JobID),
		zap.String("job_name", task.JobName),
		zap.String("execution_type", string(task.ExecutionType)),
	)

	// Initial status update: Preparing
	h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusPreparing, "Task received, preparing for execution"))

	// Create a unique workspace for this task
	workspacePath, err := os.MkdirTemp(h.cfg.WorkspaceDir, "task-"+task.JobID+"-")
	if err != nil {
		h.logger.Error("Failed to create temporary workspace", zap.String("job_id", task.JobID), zap.Error(err))
		h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusFailed, "Failed to create workspace"))
		return fmt.Errorf("failed to create workspace for job %s: %w", task.JobID, err)
	}
	defer func() {
		if err := os.RemoveAll(workspacePath); err != nil {
			h.logger.Error("Failed to clean up workspace", zap.String("job_id", task.JobID), zap.String("path", workspacePath), zap.Error(err))
		}
	}()
	h.logger.Info("Workspace created for task", zap.String("job_id", task.JobID), zap.String("path", workspacePath))

	var selectedExecutor executor.Executor
	switch task.ExecutionType {
	case models.ExecutionTypeScript:
		if h.scriptExecutor == nil {
			h.logger.Error("Script executor is not available (nil)", zap.String("job_id", task.JobID))
			h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusFailed, "Script executor not configured on this daemon"))
			return fmt.Errorf("script executor not available for job %s", task.JobID)
		}
		selectedExecutor = h.scriptExecutor
		h.logger.Info("Using ScriptExecutor for task", zap.String("job_id", task.JobID))
	case models.ExecutionTypeDocker:
		if h.dockerExecutor == nil {
			h.logger.Error("Docker executor is not available (nil), possibly due to initialization failure.", zap.String("job_id", task.JobID))
			h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusFailed, "Docker executor not available on this daemon"))
			return fmt.Errorf("docker executor not available for job %s", task.JobID)
		}
		selectedExecutor = h.dockerExecutor
		h.logger.Info("Using DockerExecutor for task", zap.String("job_id", task.JobID))
	case models.ExecutionTypeUndefined:
		h.logger.Error("Task received with undefined execution type", zap.String("job_id", task.JobID))
		h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusFailed, "Execution type not specified in task"))
		return fmt.Errorf("execution type not specified for job %s", task.JobID)
	default:
		h.logger.Error("Unsupported execution type specified", zap.String("job_id", task.JobID), zap.String("type", string(task.ExecutionType)))
		h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusFailed, fmt.Sprintf("Unsupported execution type: %s", task.ExecutionType)))
		return fmt.Errorf("unsupported execution type '%s' for job %s", task.ExecutionType, task.JobID)
	}

	// At this point, selectedExecutor is guaranteed to be non-nil if no error was returned.
	h.logger.Info("Executor selected, proceeding to execute task", zap.String("job_id", task.JobID))

	// Report StatusInProgress
	h.reporter.PublishStatus(models.NewTaskStatusUpdate(task.JobID, h.cfg.InstanceID, models.StatusInProgress, "Task execution started"))

	// 3. Execute the task using the chosen executor.
	executionResult := selectedExecutor.Execute(ctx, task, workspacePath, h.logger)

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
