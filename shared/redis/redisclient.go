package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds all configuration for the Redis client
type Config struct {
	Addr     string
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
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

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

// GetClient returns the underlying go-redis client if needed.
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

// CreateConsumerGroup creates a consumer group for the stream
func (c *RedisClient) CreateConsumerGroup(ctx context.Context, streamName, groupName string) error {
	err := c.client.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && !contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group %s: %w", groupName, err)
	}
	return nil
}

// ConsumeFromStream starts consuming messages from the stream using XREADGROUP
func (c *RedisClient) ConsumeFromStream(ctx context.Context, streamName, groupName, consumerName string) (<-chan redis.XStream, error) {
	// Create consumer group first
	if err := c.CreateConsumerGroup(ctx, streamName, groupName); err != nil {
		return nil, err
	}

	// Create channel for streaming results
	ch := make(chan redis.XStream, 1)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// XREADGROUP with blocking
				streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    groupName,
					Consumer: consumerName,
					Streams:  []string{streamName, ">"},
					Block:    0, // Block indefinitely
				}).Result()

				if err != nil {
					// Log error and retry after delay
					time.Sleep(5 * time.Second)
					continue
				}

				// Send streams to channel
				for _, stream := range streams {
					select {
					case ch <- stream:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// AcknowledgeMessage acknowledges a processed message
func (c *RedisClient) AcknowledgeMessage(ctx context.Context, streamName, groupName, messageID string) error {
	err := c.client.XAck(ctx, streamName, groupName, messageID).Err()
	if err != nil {
		return fmt.Errorf("failed to acknowledge message %s: %w", messageID, err)
	}
	return nil
}

// GetPendingMessages gets unacknowledged messages for recovery
func (c *RedisClient) GetPendingMessages(ctx context.Context, streamName, groupName string) ([]redis.XPendingExt, error) {
	result := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamName,
		Group:  groupName,
		Start:  "-",
		End:    "+",
		Count:  100,
	})

	if result.Err() != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", result.Err())
	}

	return result.Val(), nil
}

// ClaimPendingMessages claims unacknowledged messages for reprocessing
func (c *RedisClient) ClaimPendingMessages(ctx context.Context, streamName, groupName, consumerName string, minIdleTime time.Duration) ([]redis.XMessage, error) {
	result := c.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   streamName,
		Group:    groupName,
		Consumer: consumerName,
		MinIdle:  minIdleTime,
		Messages: []string{"0-0"},
	})

	if result.Err() != nil {
		return nil, fmt.Errorf("failed to claim pending messages: %w", result.Err())
	}

	return result.Val(), nil
}

// GetStreamLength returns the current stream length
func (c *RedisClient) GetStreamLength(ctx context.Context, streamName string) (int64, error) {
	return c.client.XLen(ctx, streamName).Result()
}

// HealthCheck verifies Redis connectivity
func (c *RedisClient) HealthCheck(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Helper function to check if error message contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || contains(s[1:], substr))))
}
