package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds all configuration for the Redis client
type Config struct {
	Addr     string
	Username string
	Password string
	DB       int
}

// RedisClient is a wrapper around the go-redis client.
// It provides specific methods for our audit stream.
type RedisClient struct {
	client *redis.Client
	config *Config
}

// NewClient creates and connects a new RedisClient.
func NewClient(cfg *Config) (*RedisClient, error) {
	opts := &redis.Options{
		Addr:      cfg.Addr,
		Password:  cfg.Password,
		DB:        cfg.DB,
		TLSConfig: &tls.Config{},
	}

	// Set username if provided
	if cfg.Username != "" {
		opts.Username = cfg.Username
	}

	rdb := redis.NewClient(opts)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisClient{
		client: rdb,
		config: cfg,
	}, nil
}

// Close gracefully closes the Redis connection.
func (c *RedisClient) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Redis client for advanced operations
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}

// PublishAuditEvent adds an event to the audit stream using XADD.
// 'data' should be a map[string]interface{} representing the event.
func (c *RedisClient) PublishAuditEvent(ctx context.Context, streamName string, data map[string]interface{}) (string, error) {
	// XADD streamName * key1 val1 key2 val2 ...
	// Using '*' as the ID tells Redis to auto-generate a timestamp-based ID.
	args := &redis.XAddArgs{
		Stream: streamName,
		Values: data,
	}

	msgID, err := c.client.XAdd(ctx, args).Result()
	if err != nil {
		return "", fmt.Errorf("failed to XADD to stream %s: %w", streamName, err)
	}
	return msgID, nil
}

// --- New Consumer Methods ---

// EnsureStreamGroupExists creates the consumer group (idempotent).
// Call this on consumer startup.
func (c *RedisClient) EnsureStreamGroupExists(ctx context.Context, streamName, groupName string) error {
	// XGROUP CREATE streamName groupName $ MKSTREAM
	// This command is idempotent.
	// - Creates the consumer group 'groupName' for 'streamName'.
	// - '$' means it will only read *new* messages (not historical ones).
	// - 'MKSTREAM' creates the stream if it doesn't already exist.
	err := c.client.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil {
		// "BUSYGROUP" error is fine, it means the group already exists.
		// Use string contains check instead of exact string comparison for better compatibility
		if !isBusyGroupError(err) {
			return fmt.Errorf("failed to create consumer group: %w", err)
		}
	}
	return nil
}

// ReadFromStreamGroup blocks and reads new messages from the stream.
// Returns a slice of messages or an error.
func (c *RedisClient) ReadFromStreamGroup(ctx context.Context, streamName, groupName, consumerName string, block time.Duration) ([]redis.XMessage, error) {
	// XREADGROUP GROUP groupName consumerName COUNT 1 BLOCK <block-ms> STREAMS streamName >
	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamName, ">"},
		Count:    1, // Read one at a time for safer processing
		Block:    block,
	}).Result()

	if err != nil {
		// redis.Nil indicates a timeout, which is normal
		if err == redis.Nil {
			return nil, nil // Return nil, nil to indicate no new message
		}
		return nil, fmt.Errorf("failed to XReadGroup: %w", err)
	}

	// We are only reading from one stream, so safe to return the first element
	if len(streams) > 0 {
		return streams[0].Messages, nil
	}

	return nil, nil
}

// GetPendingMessages checks for messages delivered to a consumer but not yet acknowledged.
func (c *RedisClient) GetPendingMessages(ctx context.Context, streamName, groupName, consumerName string) ([]redis.XPendingExt, error) {
	// XPENDING streamName groupName Start End Count Consumer
	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   streamName,
		Group:    groupName,
		Start:    "-", // From the beginning
		End:      "+", // To the end
		Count:    10,  // Check up to 10 pending messages
		Consumer: consumerName,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to check XPending: %w", err)
	}
	return pending, nil
}

// ClaimMessages allows a consumer to "steal" pending messages from another consumer
// (or itself) that have been idle for too long.
func (c *RedisClient) ClaimMessages(ctx context.Context, streamName, groupName, consumerName string, minIdle time.Duration, msgIDs []string) ([]redis.XMessage, error) {
	// XCLAIM stream group consumer min-idle-time msgID1 [msgID2 ...]
	claimedMsgs, err := c.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   streamName,
		Group:    groupName,
		Consumer: consumerName,
		MinIdle:  minIdle,
		Messages: msgIDs,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to XClaim messages: %w", err)
	}
	return claimedMsgs, nil
}

// AckMessage acknowledges a message as successfully processed.
func (c *RedisClient) AckMessage(ctx context.Context, streamName, groupName, msgID string) error {
	// XACK streamName groupName msgID
	err := c.client.XAck(ctx, streamName, groupName, msgID).Err()
	if err != nil {
		return fmt.Errorf("failed to XAck message %s: %w", msgID, err)
	}
	return nil
}

// isBusyGroupError checks if the error is a BUSYGROUP error indicating the consumer group already exists
func isBusyGroupError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a Redis error and contains BUSYGROUP
	if redisErr, ok := err.(redis.Error); ok {
		return strings.Contains(strings.ToUpper(redisErr.Error()), "BUSYGROUP")
	}

	// Fallback: check error message for BUSYGROUP
	return strings.Contains(strings.ToUpper(err.Error()), "BUSYGROUP")
}
