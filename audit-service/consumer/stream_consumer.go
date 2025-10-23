package consumer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/shared/redis"
)

const (
	streamName     = "audit-events"
	groupName      = "audit-processors"
	consumerName   = "audit-service-instance-1" // This should be dynamic (e.g., from hostname)
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
	client    *redis.RedisClient
	processor AuditEventProcessor
}

// NewStreamConsumer creates a new consumer and ensures the stream group exists.
func NewStreamConsumer(client *redis.RedisClient, processor AuditEventProcessor) (*StreamConsumer, error) {
	ctx := context.Background()

	// XGROUP CREATE streamName groupName $ MKSTREAM
	// This command is idempotent.
	// - Creates the consumer group 'groupName' for 'streamName'.
	// - '$' means it will only read *new* messages (not historical ones).
	// - 'MKSTREAM' creates the stream if it doesn't already exist.
	err := client.GetClient().XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil {
		// "BUSYGROUP" error is fine, it means the group already exists.
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			return nil, fmt.Errorf("failed to create consumer group: %w", err)
		}
	}
	log.Printf("Consumer group %s ensured for stream %s", groupName, streamName)

	return &StreamConsumer{
		client:    client,
		processor: processor,
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
	// XREADGROUP GROUP groupName consumerName COUNT 1 BLOCK 5000 STREAMS streamName >
	// - 'GROUP': Read as part of a consumer group.
	// - 'COUNT 1': Read one message at a time (safer for audit logs).
	// - 'BLOCK 5000': Wait up to 5 seconds for a new message.
	// - '>': Read new messages that have never been delivered to this group.
	streams, err := c.client.GetClient().XReadGroup(ctx, &redisclient.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamName, ">"},
		Count:    1,
		Block:    blockTimeout,
	}).Result()

	if err != nil {
		// Timeouts are normal, just loop again.
		if err == redisclient.Nil {
			return
		}
		log.Printf("Error in XReadGroup: %v", err)
		time.Sleep(1 * time.Second) // Avoid spamming on repeated errors
		return
	}

	// Process the received message.
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			c.processMessage(ctx, msg)
		}
	}
}

// claimPendingMessages checks for "stuck" messages and processes them.
func (c *StreamConsumer) claimPendingMessages(ctx context.Context) {
	// 1. Check for pending messages for *this* consumer
	pending, err := c.client.GetClient().XPendingExt(ctx, &redisclient.XPendingExtArgs{
		Stream:   streamName,
		Group:    groupName,
		Start:    "-", // From the beginning
		End:      "+", // To the end
		Count:    10,
		Consumer: consumerName, // Only messages delivered to *this* consumer
	}).Result()

	if err != nil {
		log.Printf("Error checking pending messages: %v", err)
		return
	}

	// 2. Claim messages that have been idle for too long
	for _, p := range pending {
		if p.Idle > pendingTimeout {
			log.Printf("Re-claiming idle message: %s", p.ID)

			// XCLAIM stream group consumer min-idle-time msgID
			claimedMsgs, err := c.client.GetClient().XClaim(ctx, &redisclient.XClaimArgs{
				Stream:   streamName,
				Group:    groupName,
				Consumer: consumerName,
				MinIdle:  pendingTimeout,
				Messages: []string{p.ID},
			}).Result()

			if err != nil {
				log.Printf("Error claiming message %s: %v", p.ID, err)
				continue
			}

			// XClaim returns the message data, process it now
			for _, msg := range claimedMsgs {
				c.processMessage(ctx, msg)
			}
		}
	}
}

// processMessage contains the core logic for processing and acknowledging.
func (c *StreamConsumer) processMessage(ctx context.Context, msg redisclient.XMessage) {
	log.Printf("Processing message: %s", msg.ID)

	// Try to process the event
	err := c.processor.ProcessAuditEvent(ctx, msg.Values)

	if err == nil {
		// SUCCESS: Acknowledge the message
		// XACK streamName groupName msgID
		if err := c.client.GetClient().XAck(ctx, streamName, groupName, msg.ID).Err(); err != nil {
			log.Printf("ERROR: Failed to XACK message %s: %v", msg.ID, err)
		}
		return
	}

	// FAILURE: Handle the error
	log.Printf("WARNING: Failed to process message %s (attempt %d): %v", msg.ID, msg.DeliveryCount, err)

	// Check if this message has failed too many times
	if msg.DeliveryCount > maxRetry {
		log.Printf("CRITICAL: Message %s has failed %d times. Moving to DLQ.", msg.ID, msg.DeliveryCount)

		// 1. Add to Dead Letter Queue (DLQ)
		dlqData := msg.Values
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
		if err := c.client.GetClient().XAck(ctx, streamName, groupName, msg.ID).Err(); err != nil {
			log.Printf("ERROR: Failed to XACK message %s after DLQ: %v", msg.ID, err)
		}
		return
	}

	// If it failed but hasn't hit retry limit, just return.
	// We don't ACK, so it will be redelivered later.
}
