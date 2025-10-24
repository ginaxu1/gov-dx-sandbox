package consumer

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gov-dx-sandbox/audit-service/redis"

	redisclient "github.com/redis/go-redis/v9"
)

// Default configuration values - can be overridden by environment variables
var (
	streamName     = getEnvOrDefault("AUDIT_STREAM_NAME", "audit-events")
	groupName      = getEnvOrDefault("AUDIT_GROUP_NAME", "audit-processors")
	dlqStreamName  = getEnvOrDefault("AUDIT_DLQ_STREAM_NAME", "audit-events_dlq")
	maxRetry       = parseIntOrDefault("AUDIT_MAX_RETRY", 5)
	blockTimeout   = parseDurationOrDefault("AUDIT_BLOCK_TIMEOUT", "5s")
	pendingTimeout = parseDurationOrDefault("AUDIT_PENDING_TIMEOUT", "1m")
)

// AuditEventProcessor defines the interface for processing a message.
// This allows us to inject our database logic easily.
type AuditEventProcessor interface {
	ProcessAuditEvent(ctx context.Context, event map[string]string) error
}

// StreamConsumer holds the logic for consuming from the Redis Stream.
type StreamConsumer struct {
	client       *redis.RedisClient
	processor    AuditEventProcessor
	consumerName string
}

// NewStreamConsumer creates a new consumer and ensures the stream group exists.
func NewStreamConsumer(client *redis.RedisClient, processor AuditEventProcessor, consumerName string) (*StreamConsumer, error) {
	ctx := context.Background()

	// Ensure the consumer group exists
	err := client.EnsureStreamGroupExists(ctx, streamName, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}
	log.Printf("Consumer group %s ensured for stream %s", groupName, streamName)

	return &StreamConsumer{
		client:       client,
		processor:    processor,
		consumerName: consumerName,
	}, nil
}

// Start consuming events in a blocking loop.
// This should be run in a goroutine from your main.go.
func (c *StreamConsumer) Start(ctx context.Context) {
	log.Println("Starting audit stream consumer...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer shutting down.")
			return
		default:
			// First, check for any old, "stuck" messages to reclaim.
			c.claimPendingMessages(ctx)

			// Now, read new messages.
			c.readNewMessages(ctx)
		}
	}
}

// readNewMessages reads new messages from the stream.
func (c *StreamConsumer) readNewMessages(ctx context.Context) {
	// Use the abstracted method from RedisClient
	messages, err := c.client.ReadFromStreamGroup(ctx, streamName, groupName, c.consumerName, blockTimeout)
	if err != nil {
		log.Printf("Error in ReadFromStreamGroup: %v", err)
		time.Sleep(1 * time.Second) // Avoid spamming on repeated errors
		return
	}

	// Process the received messages
	for _, msg := range messages {
		c.processMessage(ctx, msg)
	}
}

// claimPendingMessages checks for "stuck" messages and processes them.
func (c *StreamConsumer) claimPendingMessages(ctx context.Context) {
	// Check for pending messages for this consumer
	pending, err := c.client.GetPendingMessages(ctx, streamName, groupName, c.consumerName)
	if err != nil {
		log.Printf("Error checking pending messages: %v", err)
		return
	}

	// Claim messages that have been idle for too long
	var msgIDs []string
	for _, p := range pending {
		if p.Idle > pendingTimeout {
			log.Printf("Re-claiming idle message: %s", p.ID)
			msgIDs = append(msgIDs, p.ID)
		}
	}

	if len(msgIDs) > 0 {
		// Claim the messages
		claimedMsgs, err := c.client.ClaimMessages(ctx, streamName, groupName, c.consumerName, pendingTimeout, msgIDs)
		if err != nil {
			log.Printf("Error claiming messages: %v", err)
			return
		}

		// Process the claimed messages
		for _, msg := range claimedMsgs {
			c.processMessage(ctx, msg)
		}
	}
}

// processMessage contains the core logic for processing and acknowledging.
func (c *StreamConsumer) processMessage(ctx context.Context, msg redisclient.XMessage) {
	log.Printf("Processing message: %s", msg.ID)

	// Convert map[string]interface{} to map[string]string
	eventData := make(map[string]string)
	for k, v := range msg.Values {
		if str, ok := v.(string); ok {
			eventData[k] = str
		} else {
			// Convert non-string values to string
			eventData[k] = fmt.Sprintf("%v", v)
		}
	}

	// Get retry count from message metadata
	retryCount := c.getRetryCount(msg)

	// Try to process the event
	err := c.processor.ProcessAuditEvent(ctx, eventData)

	if err == nil {
		// SUCCESS: Acknowledge the message
		if err := c.client.AckMessage(ctx, streamName, groupName, msg.ID); err != nil {
			log.Printf("ERROR: Failed to XACK message %s: %v", msg.ID, err)
		}
		return
	}

	// FAILURE: Handle the error with retry logic
	log.Printf("WARNING: Failed to process message %s (attempt %d/%d): %v", msg.ID, retryCount+1, maxRetry, err)

	// Check if we've exceeded max retries
	if retryCount >= maxRetry {
		log.Printf("CRITICAL: Message %s exceeded max retries (%d). Moving to DLQ.", msg.ID, maxRetry)
		c.moveToDLQ(ctx, msg, err)
		return
	}

	// Increment retry count and let the message be redelivered
	log.Printf("Message %s will be retried (attempt %d/%d)", msg.ID, retryCount+1, maxRetry)
	// Don't ACK the message - let it be redelivered for retry
}

// getRetryCount extracts the retry count from message metadata
func (c *StreamConsumer) getRetryCount(msg redisclient.XMessage) int {
	if retryStr, exists := msg.Values["_retry_count"]; exists {
		if retryStr, ok := retryStr.(string); ok {
			if retryCount, err := strconv.Atoi(retryStr); err == nil {
				return retryCount
			}
		}
	}
	return 0
}

// moveToDLQ moves a failed message to the Dead Letter Queue
func (c *StreamConsumer) moveToDLQ(ctx context.Context, msg redisclient.XMessage, originalErr error) {
	// 1. Add to Dead Letter Queue (DLQ)
	dlqData := make(map[string]interface{})
	for k, v := range msg.Values {
		dlqData[k] = v
	}
	dlqData["_error"] = originalErr.Error()
	dlqData["_original_id"] = msg.ID
	dlqData["_failed_at"] = time.Now().Format(time.RFC3339)
	dlqData["_retry_count"] = c.getRetryCount(msg)

	_, dlqErr := c.client.PublishAuditEvent(ctx, dlqStreamName, dlqData)
	if dlqErr != nil {
		// This is very bad. We failed to process AND failed to DLQ.
		// The message will be redelivered, but we are stuck.
		log.Printf("FATAL: Could not move message %s to DLQ: %v", msg.ID, dlqErr)
		return // Don't ACK, let it retry.
	}

	// 2. Acknowledge the original message to remove it from the main queue
	if err := c.client.AckMessage(ctx, streamName, groupName, msg.ID); err != nil {
		log.Printf("ERROR: Failed to XACK message %s after DLQ: %v", msg.ID, err)
	}
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseIntOrDefault gets environment variable as int or returns default value
func parseIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseDurationOrDefault gets environment variable as duration or returns default value
func parseDurationOrDefault(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	if parsed, err := time.ParseDuration(defaultValue); err == nil {
		return parsed
	}
	return 5 * time.Second // fallback
}
