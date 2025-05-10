package nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dante-gpu/dante-backend/provider-daemon/internal/config"
	"github.com/dante-gpu/dante-backend/provider-daemon/internal/models"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// TaskHandlerFunc is a function type that will process received tasks.
type TaskHandlerFunc func(task *models.Task) error

// Client manages the NATS connection and subscriptions for the provider daemon.
type Client struct {
	nc           *nats.Conn
	logger       *zap.Logger
	cfg          *config.Config
	subscription *nats.Subscription
	taskHandler  TaskHandlerFunc
	shutdownChan chan struct{}
	js           nats.JetStreamContext // JetStream context
}

// NewClient creates a new NATS client for the provider daemon.
func NewClient(cfg *config.Config, logger *zap.Logger, handler TaskHandlerFunc) (*Client, error) {
	client := &Client{
		logger:       logger,
		cfg:          cfg,
		taskHandler:  handler,
		shutdownChan: make(chan struct{}),
	}

	nc, err := nats.Connect(
		cfg.NatsAddress,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(100), // More aggressive for daemon
		nats.ReconnectWait(3*time.Second),
		nats.Timeout(cfg.NatsCommandTimeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
			// Re-subscribe on reconnect if necessary (though pull consumers are more robust)
			err := client.startSubscription() // Attempt to restart subscription
			if err != nil {
				logger.Error("Failed to re-subscribe after NATS reconnect", zap.Error(err))
			}
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Warn("NATS connection closed permanently.")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logger.Error("NATS async error",
				zap.Stringp("subject", &sub.Subject),
				zap.Error(err),
			)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", cfg.NatsAddress, err)
	}
	client.nc = nc

	// Get JetStream context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close() // Close NATS connection if JS context fails
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}
	client.js = js

	logger.Info("Successfully connected to NATS and obtained JetStream context", zap.String("address", cfg.NatsAddress))
	return client, nil
}

// StartListening subscribes to the task subject and starts processing messages.
func (c *Client) StartListening() error {
	return c.startSubscription()
}

func (c *Client) startSubscription() error {
	if c.js == nil {
		return fmt.Errorf("JetStream context is not available to start subscription")
	}

	// Construct the specific subject for this daemon instance
	// The pattern from config is tasks.dispatch.%s.*, where %s is the instance ID.
	// The second '*' allows for job_id or other specific task identifiers.
	subjectToSubscribe := fmt.Sprintf(c.cfg.NatsTaskSubscriptionSubjectPattern, c.cfg.InstanceID)
	// Corrected: The pattern itself usually doesn't end with '.*' if Printf is adding it.
	// Let's assume pattern is "tasks.dispatch.%s" and we add ".*" for the subscription.
	// Or, if pattern is "tasks.dispatch.%s.*", use it directly.
	// Based on config: "tasks.dispatch.%s.*"
	// This means the scheduler will publish to tasks.dispatch.instance_id.job_id
	// The subscription subject should match this.
	// A pull consumer subscribes to a subject (which can be a wildcard) on a stream.
	// The stream should be configured to capture these subjects.

	durableName := fmt.Sprintf("provider_daemon_%s_tasks_consumer", c.cfg.InstanceID)

	c.logger.Info("Subscribing to NATS task dispatch subject (JetStream Pull)",
		zap.String("subscription_subject_pattern", c.cfg.NatsTaskSubscriptionSubjectPattern),
		zap.String("effective_subscription_subject", subjectToSubscribe),
		zap.String("durable_consumer_name", durableName),
	)

	var err error
	// Using a Pull Consumer for JetStream
	c.subscription, err = c.js.PullSubscribe(
		subjectToSubscribe, // The subject to listen on (can include wildcards like > or *)
		durableName,        // Durable name for the consumer
		nats.AckWait(c.cfg.NatsCommandTimeout*3), // AckWait should be longer than typical processing
		// nats.MaxDeliver(5), // Optional: Limit redeliveries
		// nats.BindStream("TASKS_STREAM"), // Optional: Explicitly bind to a stream if not inferred by subject
	)

	if err != nil {
		c.logger.Error("Failed to create JetStream pull subscription", zap.Error(err))
		return fmt.Errorf("failed to create pull subscription for tasks: %w", err)
	}

	go c.messageFetchLoop()
	c.logger.Info("Successfully subscribed to JetStream for task messages.")
	return nil
}

func (c *Client) messageFetchLoop() {
	c.logger.Info("Starting NATS message fetch loop...")
	batchSize := 1 // Process one task at a time for simplicity to start

	for {
		select {
		case <-c.shutdownChan:
			c.logger.Info("Shutting down NATS message fetch loop...")
			return
		default:
			if c.subscription == nil || !c.subscription.IsValid() {
				c.logger.Warn("NATS subscription is invalid or nil in fetch loop. Attempting to re-establish...")
				time.Sleep(5 * time.Second) // Wait before retrying
				if err := c.startSubscription(); err != nil {
					c.logger.Error("Failed to re-establish subscription in fetch loop", zap.Error(err))
				} else {
					c.logger.Info("Successfully re-established subscription in fetch loop.")
				}
				continue // Retry the select block
			}

			msgs, err := c.subscription.Fetch(batchSize, nats.MaxWait(10*time.Second)) // Poll with a timeout
			if err != nil {
				if err == nats.ErrTimeout {
					continue // Normal timeout, no messages
				}
				c.logger.Error("Error fetching messages from JetStream subscription", zap.Error(err))
				// Check NATS connection status
				if c.nc.Status() != nats.CONNECTED {
					c.logger.Error("NATS connection is down. Fetch loop pausing.")
					// The ReconnectHandler should attempt to re-subscribe when connection is back.
				}
				time.Sleep(2 * time.Second) // Brief pause on other errors
				continue
			}

			for _, msg := range msgs {
				c.handleMessage(msg)
			}
		}
	}
}

func (c *Client) handleMessage(msg *nats.Msg) {
	c.logger.Debug("Received raw NATS task message",
		zap.String("subject", msg.Subject),
		zap.Int("data_length", len(msg.Data)),
	)

	var task models.Task
	if err := json.Unmarshal(msg.Data, &task); err != nil {
		c.logger.Error("Failed to unmarshal task data from NATS message",
			zap.Error(err),
			zap.ByteString("raw_data", msg.Data),
		)
		if ackErr := msg.Ack(); ackErr != nil { // Ack poison pill
			c.logger.Error("Failed to ACK unmarshalable (poison pill) message", zap.Error(ackErr))
		}
		return
	}

	c.logger.Info("Successfully unmarshalled task from NATS",
		zap.String("job_id", task.JobID),
		zap.String("task_provider_id", task.AssignedProviderID),
	)

	// Ensure the task is for this specific daemon instance
	if task.AssignedProviderID != c.cfg.InstanceID {
		c.logger.Warn("Received task not assigned to this provider instance. Acknowledging and ignoring.",
			zap.String("job_id", task.JobID),
			zap.String("task_assigned_to", task.AssignedProviderID),
			zap.String("this_instance_id", c.cfg.InstanceID),
		)
		if ackErr := msg.Ack(); ackErr != nil {
			c.logger.Error("Failed to ACK misaddressed task message", zap.Error(ackErr))
		}
		return
	}

	// Process the task using the registered handler
	if err := c.taskHandler(&task); err != nil {
		c.logger.Error("Task handler failed to process task",
			zap.String("job_id", task.JobID),
			zap.Error(err),
		)
		// NAK the message to allow for retry, with a delay
		if nakErr := msg.NakWithDelay(30 * time.Second); nakErr != nil {
			c.logger.Error("Failed to NAK task message after processing error", zap.Error(nakErr))
			_ = msg.Ack() // If NAK fails, Ack to avoid loop, or Term for unrecoverable errors
		}
		return
	}

	// Task processed successfully, ACK the message
	if err := msg.AckSync(); err != nil { // AckSync waits for server confirmation
		c.logger.Error("Failed to ACK NATS message for successfully processed task",
			zap.String("job_id", task.JobID),
			zap.Error(err),
		)
	}
	c.logger.Info("Task processed and ACKed successfully", zap.String("job_id", task.JobID))
}

// PublishStatus sends a TaskStatusUpdate to the configured NATS subject.
func (c *Client) PublishStatus(statusUpdate *models.TaskStatusUpdate) error {
	if c.nc == nil || c.nc.Status() != nats.CONNECTED {
		return fmt.Errorf("NATS client not connected, cannot publish status for job %s", statusUpdate.JobID)
	}

	statusJSON, err := json.Marshal(statusUpdate)
	if err != nil {
		c.logger.Error("Failed to marshal task status update", zap.String("job_id", statusUpdate.JobID), zap.Error(err))
		return fmt.Errorf("failed to marshal status update: %w", err)
	}

	subject := fmt.Sprintf("%s.%s", c.cfg.NatsJobStatusUpdateSubjectPrefix, statusUpdate.JobID)
	c.logger.Debug("Publishing task status update",
		zap.String("subject", subject),
		zap.String("job_id", statusUpdate.JobID),
		zap.String("status", string(statusUpdate.Status)),
	)

	if err := c.nc.Publish(subject, statusJSON); err != nil {
		c.logger.Error("Failed to publish task status update to NATS", zap.String("subject", subject), zap.Error(err))
		return fmt.Errorf("failed to publish status update to NATS: %w", err)
	}
	// For important status updates, ensure it reaches the server.
	// Use PublishMsg with reply or ensure JetStream acks if publishing to a stream.
	// For plain publish, Flush or FlushTimeout helps.
	return c.nc.FlushTimeout(c.cfg.NatsCommandTimeout)
}

// Stop gracefully shuts down the NATS client.
func (c *Client) Stop() {
	c.logger.Info("Stopping NATS client for provider daemon...")
	close(c.shutdownChan) // Signal messageFetchLoop to stop

	if c.subscription != nil && c.subscription.IsValid() {
		c.logger.Info("Draining NATS JetStream subscription...")
		if err := c.subscription.Drain(); err != nil {
			c.logger.Error("Error draining NATS JetStream subscription", zap.Error(err))
		} else {
			c.logger.Info("NATS JetStream subscription drained.")
		}
		// Unsubscribe is not strictly necessary after Drain for pull consumers if the consumer is deleted,
		// but doesn't hurt if the durable consumer is meant to persist.
		// If we want to remove the durable consumer, it's a different API call (js.DeleteConsumer).
	}

	if c.nc != nil {
		c.logger.Info("Draining NATS connection...")
		if err := c.nc.Drain(); err != nil {
			c.logger.Error("Error draining NATS connection", zap.Error(err))
		} else {
			c.logger.Info("NATS connection drained.")
		}
		c.nc.Close() // Close the connection
		c.logger.Info("NATS connection closed.")
	}
	c.logger.Info("NATS client stopped.")
}
