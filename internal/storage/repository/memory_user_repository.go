package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrAPIKeyNotFound    = errors.New("API key not found")
)

// MemoryUserRepository implements UserRepository with in-memory storage
type MemoryUserRepository struct {
	users   map[string]*userWithPassword
	byEmail map[string]string // email -> id
	byName  map[string]string // username -> id
	mu      sync.RWMutex
}

type userWithPassword struct {
	user         *auth.User
	passwordHash string
}

// NewMemoryUserRepository creates a new in-memory user repository with default admin
func NewMemoryUserRepository() *MemoryUserRepository {
	repo := &MemoryUserRepository{
		users:   make(map[string]*userWithPassword),
		byEmail: make(map[string]string),
		byName:  make(map[string]string),
	}

	// Create default admin user (password: admin123)
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := &auth.User{
		ID:        "user-admin-001",
		Username:  "admin",
		Email:     "admin@volcanion.com",
		Role:      auth.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.users[adminUser.ID] = &userWithPassword{
		user:         adminUser,
		passwordHash: string(adminHash),
	}
	repo.byEmail[adminUser.Email] = adminUser.ID
	repo.byName[adminUser.Username] = adminUser.ID

	return repo
}

func (r *MemoryUserRepository) Create(user *auth.User, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return ErrUserAlreadyExists
	}
	if _, exists := r.byEmail[user.Email]; exists {
		return ErrUserAlreadyExists
	}
	if _, exists := r.byName[user.Username]; exists {
		return ErrUserAlreadyExists
	}

	r.users[user.ID] = &userWithPassword{
		user:         user,
		passwordHash: passwordHash,
	}
	r.byEmail[user.Email] = user.ID
	r.byName[user.Username] = user.ID
	return nil
}

func (r *MemoryUserRepository) GetByID(id string) (*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	uwp, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return uwp.user, nil
}

func (r *MemoryUserRepository) GetByUsername(username string) (*auth.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byName[username]
	if !exists {
		return nil, "", ErrUserNotFound
	}
	uwp := r.users[id]
	return uwp.user, uwp.passwordHash, nil
}

func (r *MemoryUserRepository) GetByEmail(email string) (*auth.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byEmail[email]
	if !exists {
		return nil, "", ErrUserNotFound
	}
	uwp := r.users[id]
	return uwp.user, uwp.passwordHash, nil
}

func (r *MemoryUserRepository) Update(user *auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	uwp, exists := r.users[user.ID]
	if !exists {
		return ErrUserNotFound
	}

	// Update email index if changed
	if uwp.user.Email != user.Email {
		delete(r.byEmail, uwp.user.Email)
		r.byEmail[user.Email] = user.ID
	}

	// Update username index if changed
	if uwp.user.Username != user.Username {
		delete(r.byName, uwp.user.Username)
		r.byName[user.Username] = user.ID
	}

	uwp.user = user
	return nil
}

func (r *MemoryUserRepository) UpdatePassword(userID string, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	uwp, exists := r.users[userID]
	if !exists {
		return ErrUserNotFound
	}
	uwp.passwordHash = passwordHash
	return nil
}

func (r *MemoryUserRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	uwp, exists := r.users[id]
	if !exists {
		return ErrUserNotFound
	}

	delete(r.byEmail, uwp.user.Email)
	delete(r.byName, uwp.user.Username)
	delete(r.users, id)
	return nil
}

func (r *MemoryUserRepository) List() ([]*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*auth.User, 0, len(r.users))
	for _, uwp := range r.users {
		users = append(users, uwp.user)
	}
	return users, nil
}

// MemoryAPIKeyRepository implements APIKeyRepository with in-memory storage
type MemoryAPIKeyRepository struct {
	keys     map[string]*auth.APIKey
	byHash   map[string]string   // keyHash -> id
	byUserID map[string][]string // userID -> []id
	mu       sync.RWMutex
}

// NewMemoryAPIKeyRepository creates a new in-memory API key repository
func NewMemoryAPIKeyRepository() *MemoryAPIKeyRepository {
	return &MemoryAPIKeyRepository{
		keys:     make(map[string]*auth.APIKey),
		byHash:   make(map[string]string),
		byUserID: make(map[string][]string),
	}
}

func (r *MemoryAPIKeyRepository) Create(apiKey *auth.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.keys[apiKey.ID] = apiKey
	r.byHash[apiKey.Key] = apiKey.ID
	r.byUserID[apiKey.UserID] = append(r.byUserID[apiKey.UserID], apiKey.ID)
	return nil
}

func (r *MemoryAPIKeyRepository) GetByID(id string) (*auth.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key, exists := r.keys[id]
	if !exists {
		return nil, ErrAPIKeyNotFound
	}
	return key, nil
}

func (r *MemoryAPIKeyRepository) GetByKeyHash(keyHash string) (*auth.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byHash[keyHash]
	if !exists {
		return nil, ErrAPIKeyNotFound
	}
	return r.keys[id], nil
}

func (r *MemoryAPIKeyRepository) GetByUserID(userID string) ([]*auth.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byUserID[userID]
	keys := make([]*auth.APIKey, 0, len(ids))
	for _, id := range ids {
		if key, exists := r.keys[id]; exists {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (r *MemoryAPIKeyRepository) Update(apiKey *auth.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.keys[apiKey.ID]; !exists {
		return ErrAPIKeyNotFound
	}
	r.keys[apiKey.ID] = apiKey
	return nil
}

func (r *MemoryAPIKeyRepository) UpdateLastUsed(id string, lastUsed time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key, exists := r.keys[id]
	if !exists {
		return ErrAPIKeyNotFound
	}
	key.LastUsed = &lastUsed
	return nil
}

func (r *MemoryAPIKeyRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key, exists := r.keys[id]
	if !exists {
		return ErrAPIKeyNotFound
	}

	delete(r.byHash, key.Key)
	delete(r.keys, id)

	// Remove from user's key list
	userKeys := r.byUserID[key.UserID]
	for i, kid := range userKeys {
		if kid == id {
			r.byUserID[key.UserID] = append(userKeys[:i], userKeys[i+1:]...)
			break
		}
	}
	return nil
}

func (r *MemoryAPIKeyRepository) Revoke(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key, exists := r.keys[id]
	if !exists {
		return ErrAPIKeyNotFound
	}
	key.IsActive = false
	return nil
}

// MemoryTokenBlacklistRepository implements TokenBlacklistRepository
type MemoryTokenBlacklistRepository struct {
	tokens map[string]tokenEntry
	mu     sync.RWMutex
}

type tokenEntry struct {
	userID    string
	expiresAt time.Time
}

// NewMemoryTokenBlacklistRepository creates a new in-memory token blacklist
func NewMemoryTokenBlacklistRepository() *MemoryTokenBlacklistRepository {
	return &MemoryTokenBlacklistRepository{
		tokens: make(map[string]tokenEntry),
	}
}

func (r *MemoryTokenBlacklistRepository) Add(tokenHash string, userID string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[tokenHash] = tokenEntry{userID: userID, expiresAt: expiresAt}
	return nil
}

func (r *MemoryTokenBlacklistRepository) Exists(tokenHash string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.tokens[tokenHash]
	return exists, nil
}

func (r *MemoryTokenBlacklistRepository) CleanExpired() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for hash, entry := range r.tokens {
		if entry.expiresAt.Before(now) {
			delete(r.tokens, hash)
		}
	}
	return nil
}
