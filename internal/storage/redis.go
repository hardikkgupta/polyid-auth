package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCache implements caching using Redis
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, logger *zap.Logger) *RedisCache {
	return &RedisCache{
		client: client,
		logger: logger,
	}
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", &StorageError{
			Code:    ErrNotFound,
			Message: "Cache key not found",
		}
	}
	if err != nil {
		return "", &StorageError{
			Code:    ErrInternal,
			Message: "Failed to get from cache",
			Err:     err,
		}
	}
	return val, nil
}

// Set stores a value in the cache with an expiration
func (c *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := c.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to set cache value",
			Err:     err,
		}
	}
	return nil
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to delete cache value",
			Err:     err,
		}
	}
	return nil
}

// GetUser retrieves a user from the cache
func (c *RedisCache) GetUser(ctx context.Context, userID string) (*User, error) {
	key := fmt.Sprintf("user:%s", userID)
	data, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to unmarshal user from cache",
			Err:     err,
		}
	}

	return &user, nil
}

// SetUser stores a user in the cache
func (c *RedisCache) SetUser(ctx context.Context, user *User, expiration time.Duration) error {
	key := fmt.Sprintf("user:%s", user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to marshal user for cache",
			Err:     err,
		}
	}

	return c.Set(ctx, key, string(data), expiration)
}

// GetCredentials retrieves credentials from the cache
func (c *RedisCache) GetCredentials(ctx context.Context, userID string) ([]*Credential, error) {
	key := fmt.Sprintf("credentials:%s", userID)
	data, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var credentials []*Credential
	if err := json.Unmarshal([]byte(data), &credentials); err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to unmarshal credentials from cache",
			Err:     err,
		}
	}

	return credentials, nil
}

// SetCredentials stores credentials in the cache
func (c *RedisCache) SetCredentials(ctx context.Context, userID string, credentials []*Credential, expiration time.Duration) error {
	key := fmt.Sprintf("credentials:%s", userID)
	data, err := json.Marshal(credentials)
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to marshal credentials for cache",
			Err:     err,
		}
	}

	return c.Set(ctx, key, string(data), expiration)
}

// GetMFAMethods retrieves MFA methods from the cache
func (c *RedisCache) GetMFAMethods(ctx context.Context, userID string) ([]*MFAMethod, error) {
	key := fmt.Sprintf("mfa:%s", userID)
	data, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var methods []*MFAMethod
	if err := json.Unmarshal([]byte(data), &methods); err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to unmarshal MFA methods from cache",
			Err:     err,
		}
	}

	return methods, nil
}

// SetMFAMethods stores MFA methods in the cache
func (c *RedisCache) SetMFAMethods(ctx context.Context, userID string, methods []*MFAMethod, expiration time.Duration) error {
	key := fmt.Sprintf("mfa:%s", userID)
	data, err := json.Marshal(methods)
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to marshal MFA methods for cache",
			Err:     err,
		}
	}

	return c.Set(ctx, key, string(data), expiration)
}

// InvalidateUser invalidates all user-related cache entries
func (c *RedisCache) InvalidateUser(ctx context.Context, userID string) error {
	patterns := []string{
		fmt.Sprintf("user:%s", userID),
		fmt.Sprintf("credentials:%s", userID),
		fmt.Sprintf("mfa:%s", userID),
	}

	for _, pattern := range patterns {
		keys, err := c.client.Keys(ctx, pattern).Result()
		if err != nil {
			return &StorageError{
				Code:    ErrInternal,
				Message: "Failed to get cache keys",
				Err:     err,
			}
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return &StorageError{
					Code:    ErrInternal,
					Message: "Failed to delete cache keys",
					Err:     err,
				}
			}
		}
	}

	return nil
} 