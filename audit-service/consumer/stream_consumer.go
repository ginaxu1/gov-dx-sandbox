package consumer

import (
	"context"
	"fmt"
	"log"
	"time"

	redisclient "github.com/gov-dx-sandbox/shared/redis"
	"github.com/redis/go-redis/v9"
)

const (
	streamName     = "audit-events"
	groupName      = "audit-processors"
	dlqStreamName  = "audit-events_dlq"
	maxRetry       = 5
	blockTimeout   = 5 * time.Second
	pendingTimeout = 1 * time.Minute // Time before a message is considered "stuck"
)

// AuditEventProcessor defines the interface for processing a message.
// This allows us to inject our database logic easily.
type AuditEventProcessor interface {
	ProcessAuditEvent(ctx context.Context, event map[string]string) error
}

// StreamConsumer holds the logic for consuming from the Redis Stream.
type StreamConsumer struct {
	client       *redisclient.RedisClient
	processor    AuditEventProcessor
	consumerName string
}

// NewStreamConsumer creates a new consumer and ensures the stream group exists.
func NewStreamConsumer(client *redisclient.RedisClient, processor AuditEventProcessor, consumerName string) (*StreamConsumer, error) {
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
func (c *StreamConsumer) processMessage(ctx context.Context, msg redis.XMessage) {
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

	// Try to process the event
	err := c.processor.ProcessAuditEvent(ctx, eventData)

	if err == nil {
		// SUCCESS: Acknowledge the message
		if err := c.client.AckMessage(ctx, streamName, groupName, msg.ID); err != nil {
			log.Printf("ERROR: Failed to XACK message %s: %v", msg.ID, err)
		}
		return
	}

	// FAILURE: Handle the error
	log.Printf("WARNING: Failed to process message %s: %v", msg.ID, err)

	// For now, we'll move failed messages to DLQ immediately
	// In a production system, you'd want to track retry counts
	log.Printf("CRITICAL: Message %s failed. Moving to DLQ.", msg.ID)

	// 1. Add to Dead Letter Queue (DLQ)
	dlqData := make(map[string]interface{})
	for k, v := range msg.Values {
		dlqData[k] = v
	}
	dlqData["_error"] = err.Error()
	dlqData["_original_id"] = msg.ID
	dlqData["_failed_at"] = time.Now().Format(time.RFC3339)

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
