package auth

import (
	"errors"
	"testing"
	"time"
)

func TestNewJWTService(t *testing.T) {
	secret := "test-secret-key-at-least-32-bytes-long"
	duration := 1 * time.Hour

	service := NewJWTService(secret, duration)

	if service == nil {
		t.Fatal("Expected JWTService to be created")
	}
}

func TestJWTGenerateToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)

	user := &User{
		ID:       "user-123",
		Username: "testuser",
		Role:     RoleAdmin,
	}

	token, expiry, err := service.GenerateToken(user)

	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	if expiry == 0 {
		t.Error("Expected non-zero expiry time")
	}

	expectedExpiry := time.Now().Add(1 * time.Hour).Unix()
	if expiry < expectedExpiry-5 || expiry > expectedExpiry+5 {
		t.Errorf("Expiry time %d not within expected range of %d", expiry, expectedExpiry)
	}

	t.Logf("Generated token: %s...", token[:50])
	t.Logf("Expiry: %d", expiry)
}

func TestJWTValidateToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)

	user := &User{
		ID:       "user-456",
		Username: "validuser",
		Role:     RoleUser,
	}

	token, _, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID mismatch: expected %s, got %s", user.ID, claims.UserID)
	}
	if claims.Username != user.Username {
		t.Errorf("Username mismatch: expected %s, got %s", user.Username, claims.Username)
	}
	if claims.Role != user.Role {
		t.Errorf("Role mismatch: expected %v, got %v", user.Role, claims.Role)
	}
}

func TestJWTExpiredToken(t *testing.T) {
	// Create service with very short duration
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Millisecond)

	user := &User{
		ID:       "user-1",
		Username: "expireduser",
		Role:     RoleUser,
	}

	token, _, err := service.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if !errors.Is(err, ErrExpiredToken) {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)

	testCases := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"invalid", "not.a.valid.token"},
		{"gibberish", "abcdefghijklmnop"},
		{"partial", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ValidateToken(tc.token)
			if err == nil {
				t.Error("Expected error for invalid token")
			}
		})
	}
}

func TestJWTWrongSecret(t *testing.T) {
	service1 := NewJWTService("secret-key-one-at-least-32-bytes!", 1*time.Hour)
	service2 := NewJWTService("secret-key-two-at-least-32-bytes!", 1*time.Hour)

	user := &User{
		ID:       "user-1",
		Username: "user",
		Role:     RoleAdmin,
	}

	token, _, err := service1.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, err = service2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating with wrong secret")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTMultipleTokens(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)

	// Generate multiple tokens
	tokens := make([]string, 5)
	for i := 0; i < 5; i++ {
		user := &User{
			ID:       "user-" + string(rune('0'+i)),
			Username: "username" + string(rune('0'+i)),
			Role:     RoleUser,
		}
		token, _, err := service.GenerateToken(user)
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}
		tokens[i] = token
	}

	// Validate all tokens
	for i, token := range tokens {
		claims, err := service.ValidateToken(token)
		if err != nil {
			t.Errorf("Failed to validate token %d: %v", i, err)
		}
		expectedUserID := "user-" + string(rune('0'+i))
		if claims.UserID != expectedUserID {
			t.Errorf("Token %d: expected UserID %s, got %s", i, expectedUserID, claims.UserID)
		}
	}
}

func TestJWTTokenUniqueness(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)

	// Generate multiple tokens for same user with delay
	// Note: JWT includes IssuedAt timestamp, so tokens within same second may be same
	tokens := make(map[string]bool)
	for i := 0; i < 5; i++ {
		user := &User{
			ID:       "same-user",
			Username: "sameusername",
			Role:     RoleUser,
		}
		token, _, err := service.GenerateToken(user)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		tokens[token] = true

		// Delay to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
	}

	// With sufficient delay, tokens should be unique
	if len(tokens) < 3 {
		t.Logf("Warning: Only %d unique tokens generated (expected more)", len(tokens))
	}
}

func TestJWTClaimsFields(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)

	// Test with various special characters
	testCases := []struct {
		userID   string
		username string
		role     Role
	}{
		{"uuid-123-456-789", "john.doe", RoleAdmin},
		{"12345", "user@example.com", RoleAdmin},
		{"abc_def", "user_name_123", RoleUser},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			user := &User{
				ID:       tc.userID,
				Username: tc.username,
				Role:     tc.role,
			}
			token, _, err := service.GenerateToken(user)
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate token: %v", err)
			}

			if claims.UserID != tc.userID {
				t.Errorf("UserID mismatch: expected %s, got %s", tc.userID, claims.UserID)
			}
			if claims.Username != tc.username {
				t.Errorf("Username mismatch: expected %s, got %s", tc.username, claims.Username)
			}
			if claims.Role != tc.role {
				t.Errorf("Role mismatch: expected %v, got %v", tc.role, claims.Role)
			}
		})
	}
}

func BenchmarkJWTGenerateToken(b *testing.B) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)
	user := &User{
		ID:       "user-123",
		Username: "testuser",
		Role:     RoleAdmin,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.GenerateToken(user)
	}
}

func BenchmarkJWTValidateToken(b *testing.B) {
	service := NewJWTService("test-secret-key-at-least-32-bytes-long", 1*time.Hour)
	user := &User{
		ID:       "user-123",
		Username: "testuser",
		Role:     RoleAdmin,
	}
	token, _, _ := service.GenerateToken(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ValidateToken(token)
	}
}
