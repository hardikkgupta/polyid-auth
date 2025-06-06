package storage

import (
	"context"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Credential represents a WebAuthn credential
type Credential struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	PublicKey       []byte    `json:"public_key"`
	AttestationType string    `json:"attestation_type"`
	CreatedAt       time.Time `json:"created_at"`
}

// MFAMethod represents a user's MFA method
type MFAMethod struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // "totp", "sms", "app_link"
	Value     string    `json:"value"` // secret, phone number, or device ID
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Storage defines the interface for data persistence
type Storage interface {
	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error

	// Credential operations
	StoreCredential(ctx context.Context, credential *Credential) error
	GetCredentials(ctx context.Context, userID string) ([]*Credential, error)
	DeleteCredential(ctx context.Context, id string) error

	// MFA operations
	StoreMFAMethod(ctx context.Context, method *MFAMethod) error
	GetMFAMethods(ctx context.Context, userID string) ([]*MFAMethod, error)
	DeleteMFAMethod(ctx context.Context, id string) error

	// Temporary storage operations (for verification flows)
	StoreTemporaryValue(ctx context.Context, key string, value string, expiry time.Duration) error
	GetTemporaryValue(ctx context.Context, key string) (string, error)
	DeleteTemporaryValue(ctx context.Context, key string) error

	// Session operations
	StoreSession(ctx context.Context, sessionID string, userID string, expiry time.Duration) error
	GetSession(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

// StorageError represents a storage-specific error
type StorageError struct {
	Code    string
	Message string
	Err     error
}

func (e *StorageError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Common error codes
const (
	ErrNotFound     = "NOT_FOUND"
	ErrAlreadyExists = "ALREADY_EXISTS"
	ErrInvalidInput = "INVALID_INPUT"
	ErrInternal     = "INTERNAL_ERROR"
) 