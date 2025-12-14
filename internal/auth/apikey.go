package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrAPIKeyExpired  = errors.New("API key has expired")
	ErrAPIKeyInactive = errors.New("API key is inactive")
)

// APIKeyService handles API key operations
type APIKeyService struct {
	keys map[string]*APIKey // Map of hashed key -> APIKey
	mu   sync.RWMutex
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService() *APIKeyService {
	return &APIKeyService{
		keys: make(map[string]*APIKey),
	}
}

// CreateAPIKey creates a new API key
func (s *APIKeyService) CreateAPIKey(userID string, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	// Generate random API key
	plainKey, err := GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	// Hash the key for storage
	hashedKey := hashAPIKey(plainKey)

	apiKey := &APIKey{
		ID:        uuid.New().String(),
		Key:       hashedKey,
		Name:      req.Name,
		UserID:    userID,
		Role:      req.Role,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	s.mu.Lock()
	s.keys[hashedKey] = apiKey
	s.mu.Unlock()

	return &CreateAPIKeyResponse{
		ID:        apiKey.ID,
		Key:       plainKey, // Return plaintext key (only time it's visible)
		Name:      apiKey.Name,
		Role:      apiKey.Role,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
	}, nil
}

// ValidateAPIKey validates an API key and returns the associated key info
func (s *APIKeyService) ValidateAPIKey(plainKey string) (*APIKey, error) {
	hashedKey := hashAPIKey(plainKey)

	s.mu.RLock()
	apiKey, exists := s.keys[hashedKey]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrAPIKeyNotFound
	}

	if !apiKey.IsActive {
		return nil, ErrAPIKeyInactive
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}

	// Update last used timestamp
	s.mu.Lock()
	now := time.Now()
	apiKey.LastUsed = &now
	s.mu.Unlock()

	return apiKey, nil
}

// RevokeAPIKey deactivates an API key
func (s *APIKeyService) RevokeAPIKey(keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, apiKey := range s.keys {
		if apiKey.ID == keyID {
			apiKey.IsActive = false
			return nil
		}
	}

	return ErrAPIKeyNotFound
}

// ListAPIKeys returns all API keys for a user
func (s *APIKeyService) ListAPIKeys(userID string) []*APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userKeys []*APIKey
	for _, apiKey := range s.keys {
		if apiKey.UserID == userID {
			// Return a copy without the hashed key
			keyCopy := *apiKey
			keyCopy.Key = "" // Don't expose hashed key
			userKeys = append(userKeys, &keyCopy)
		}
	}

	return userKeys
}

// hashAPIKey hashes an API key using SHA-256
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
