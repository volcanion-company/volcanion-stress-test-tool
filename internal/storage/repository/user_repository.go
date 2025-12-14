package repository

import (
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/auth"
)

// UserRepository interface for user data access
type UserRepository interface {
	Create(user *auth.User, passwordHash string) error
	GetByID(id string) (*auth.User, error)
	GetByUsername(username string) (*auth.User, string, error) // Returns user and password hash
	GetByEmail(email string) (*auth.User, string, error)       // Returns user and password hash
	Update(user *auth.User) error
	UpdatePassword(userID string, passwordHash string) error
	Delete(id string) error
	List() ([]*auth.User, error)
}

// APIKeyRepository interface for API key data access
type APIKeyRepository interface {
	Create(apiKey *auth.APIKey) error
	GetByID(id string) (*auth.APIKey, error)
	GetByKeyHash(keyHash string) (*auth.APIKey, error)
	GetByUserID(userID string) ([]*auth.APIKey, error)
	Update(apiKey *auth.APIKey) error
	UpdateLastUsed(id string, lastUsed time.Time) error
	Delete(id string) error
	Revoke(id string) error
}

// TokenBlacklistRepository interface for token blacklist
type TokenBlacklistRepository interface {
	Add(tokenHash string, userID string, expiresAt time.Time) error
	Exists(tokenHash string) (bool, error)
	CleanExpired() error
}
