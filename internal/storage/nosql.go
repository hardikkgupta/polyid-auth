package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// NoSQLStorage implements the Storage interface using a generic NoSQL database
type NoSQLStorage struct {
	client    NoSQLClient
	logger    *zap.Logger
	tableName string
}

// NoSQLClient defines the interface for NoSQL database operations
type NoSQLClient interface {
	Put(ctx context.Context, table string, key string, value interface{}) error
	Get(ctx context.Context, table string, key string) (map[string]interface{}, error)
	Query(ctx context.Context, table string, index string, condition string, params map[string]interface{}) ([]map[string]interface{}, error)
	Delete(ctx context.Context, table string, key string) error
	CreateIndex(ctx context.Context, table string, index string, fields []string) error
}

// NewNoSQLStorage creates a new NoSQL storage instance
func NewNoSQLStorage(client NoSQLClient, logger *zap.Logger, tableName string) *NoSQLStorage {
	return &NoSQLStorage{
		client:    client,
		logger:    logger,
		tableName: tableName,
	}
}

// CreateUser implements Storage.CreateUser
func (s *NoSQLStorage) CreateUser(ctx context.Context, user *User) error {
	// Check if user already exists
	_, err := s.client.Get(ctx, s.tableName, user.ID)
	if err == nil {
		return &StorageError{
			Code:    ErrAlreadyExists,
			Message: "User already exists",
		}
	}

	// Create user
	err = s.client.Put(ctx, s.tableName, user.ID, user)
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to create user",
			Err:     err,
		}
	}

	return nil
}

// GetUser implements Storage.GetUser
func (s *NoSQLStorage) GetUser(ctx context.Context, id string) (*User, error) {
	result, err := s.client.Get(ctx, s.tableName, id)
	if err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to get user",
			Err:     err,
		}
	}

	if result == nil {
		return nil, &StorageError{
			Code:    ErrNotFound,
			Message: "User not found",
		}
	}

	user := &User{}
	err = mapToStruct(result, user)
	if err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to unmarshal user",
			Err:     err,
		}
	}

	return user, nil
}

// GetUserByEmail implements Storage.GetUserByEmail
func (s *NoSQLStorage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	results, err := s.client.Query(ctx, s.tableName, "email-index", "email = :email", map[string]interface{}{
		":email": email,
	})
	if err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to query user by email",
			Err:     err,
		}
	}

	if len(results) == 0 {
		return nil, &StorageError{
			Code:    ErrNotFound,
			Message: "User not found",
		}
	}

	user := &User{}
	err = mapToStruct(results[0], user)
	if err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to unmarshal user",
			Err:     err,
		}
	}

	return user, nil
}

// StoreCredential implements Storage.StoreCredential
func (s *NoSQLStorage) StoreCredential(ctx context.Context, credential *Credential) error {
	err := s.client.Put(ctx, s.tableName, credential.ID, credential)
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to store credential",
			Err:     err,
		}
	}

	return nil
}

// GetCredentials implements Storage.GetCredentials
func (s *NoSQLStorage) GetCredentials(ctx context.Context, userID string) ([]*Credential, error) {
	results, err := s.client.Query(ctx, s.tableName, "user-credentials-index", "user_id = :user_id", map[string]interface{}{
		":user_id": userID,
	})
	if err != nil {
		return nil, &StorageError{
			Code:    ErrInternal,
			Message: "Failed to query credentials",
			Err:     err,
		}
	}

	credentials := make([]*Credential, 0, len(results))
	for _, result := range results {
		credential := &Credential{}
		err := mapToStruct(result, credential)
		if err != nil {
			return nil, &StorageError{
				Code:    ErrInternal,
				Message: "Failed to unmarshal credential",
				Err:     err,
			}
		}
		credentials = append(credentials, credential)
	}

	return credentials, nil
}

// StoreTemporaryValue implements Storage.StoreTemporaryValue
func (s *NoSQLStorage) StoreTemporaryValue(ctx context.Context, key string, value string, expiry time.Duration) error {
	tempValue := map[string]interface{}{
		"value":     value,
		"expires_at": time.Now().Add(expiry).Unix(),
	}

	err := s.client.Put(ctx, s.tableName, fmt.Sprintf("temp:%s", key), tempValue)
	if err != nil {
		return &StorageError{
			Code:    ErrInternal,
			Message: "Failed to store temporary value",
			Err:     err,
		}
	}

	return nil
}

// GetTemporaryValue implements Storage.GetTemporaryValue
func (s *NoSQLStorage) GetTemporaryValue(ctx context.Context, key string) (string, error) {
	result, err := s.client.Get(ctx, s.tableName, fmt.Sprintf("temp:%s", key))
	if err != nil {
		return "", &StorageError{
			Code:    ErrInternal,
			Message: "Failed to get temporary value",
			Err:     err,
		}
	}

	if result == nil {
		return "", &StorageError{
			Code:    ErrNotFound,
			Message: "Temporary value not found",
		}
	}

	// Check expiration
	expiresAt, ok := result["expires_at"].(float64)
	if !ok {
		return "", &StorageError{
			Code:    ErrInternal,
			Message: "Invalid expiration time",
		}
	}

	if time.Now().Unix() > int64(expiresAt) {
		// Delete expired value
		_ = s.client.Delete(ctx, s.tableName, fmt.Sprintf("temp:%s", key))
		return "", &StorageError{
			Code:    ErrNotFound,
			Message: "Temporary value expired",
		}
	}

	value, ok := result["value"].(string)
	if !ok {
		return "", &StorageError{
			Code:    ErrInternal,
			Message: "Invalid value type",
		}
	}

	return value, nil
}

// Helper function to convert map to struct
func mapToStruct(m map[string]interface{}, v interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
} 