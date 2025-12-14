package auth

import (
	"testing"
	"time"
)

func TestNewAPIKeyService(t *testing.T) {
	service := NewAPIKeyService()

	if service == nil {
		t.Fatal("Expected APIKeyService to be created")
	}
}

func TestAPIKeyCreate(t *testing.T) {
	service := NewAPIKeyService()

	name := "Test API Key"
	userID := "user-123"
	expiry := time.Now().Add(24 * time.Hour)

	req := &CreateAPIKeyRequest{
		Name:      name,
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}

	resp, err := service.CreateAPIKey(userID, req)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	if resp.Key == "" {
		t.Error("Expected plain text key to be returned")
	}
	if resp.Name != name {
		t.Errorf("Name mismatch: expected %s, got %s", name, resp.Name)
	}
	if resp.ID == "" {
		t.Error("Expected ID to be generated")
	}

	t.Logf("Created API key: %s... (ID: %s)", resp.Key[:20], resp.ID)
}

func TestAPIKeyValidate(t *testing.T) {
	service := NewAPIKeyService()

	name := "Validation Test Key"
	userID := "user-456"
	expiry := time.Now().Add(24 * time.Hour)

	req := &CreateAPIKeyRequest{
		Name:      name,
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}

	resp, err := service.CreateAPIKey(userID, req)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// Validate with plain text key
	validatedKey, err := service.ValidateAPIKey(resp.Key)
	if err != nil {
		t.Fatalf("Failed to validate API key: %v", err)
	}

	if validatedKey.ID != resp.ID {
		t.Errorf("ID mismatch: expected %s, got %s", resp.ID, validatedKey.ID)
	}
	if validatedKey.UserID != userID {
		t.Errorf("UserID mismatch: expected %s, got %s", userID, validatedKey.UserID)
	}
}

func TestAPIKeyInvalidKey(t *testing.T) {
	service := NewAPIKeyService()

	testCases := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"invalid", "not-a-valid-key"},
		{"random", "abcdef123456"},
		{"uuid-like", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ValidateAPIKey(tc.key)
			if err == nil {
				t.Error("Expected error for invalid key")
			}
		})
	}
}

func TestAPIKeyExpired(t *testing.T) {
	service := NewAPIKeyService()

	// Create key that expires immediately
	expiry := time.Now().Add(-1 * time.Hour) // Already expired

	req := &CreateAPIKeyRequest{
		Name:      "Expired Key",
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}

	resp, err := service.CreateAPIKey("user-1", req)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	_, err = service.ValidateAPIKey(resp.Key)
	if err == nil {
		t.Error("Expected error for expired key")
	}
	if err != ErrAPIKeyExpired {
		t.Errorf("Expected ErrAPIKeyExpired, got %v", err)
	}
}

func TestAPIKeyRevoke(t *testing.T) {
	service := NewAPIKeyService()

	expiry := time.Now().Add(24 * time.Hour)
	req := &CreateAPIKeyRequest{
		Name:      "Revoke Test Key",
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}

	resp, err := service.CreateAPIKey("user-1", req)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// Key should be valid initially
	_, err = service.ValidateAPIKey(resp.Key)
	if err != nil {
		t.Fatalf("Key should be valid before revocation: %v", err)
	}

	// Revoke the key
	err = service.RevokeAPIKey(resp.ID)
	if err != nil {
		t.Fatalf("Failed to revoke API key: %v", err)
	}

	// Key should be invalid after revocation
	_, err = service.ValidateAPIKey(resp.Key)
	if err == nil {
		t.Error("Expected error for revoked key")
	}
	if err != ErrAPIKeyInactive {
		t.Errorf("Expected ErrAPIKeyInactive, got %v", err)
	}
}

func TestAPIKeyRevokeNonExistent(t *testing.T) {
	service := NewAPIKeyService()

	err := service.RevokeAPIKey("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
	if err != ErrAPIKeyNotFound {
		t.Errorf("Expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestAPIKeyLastUsedUpdate(t *testing.T) {
	service := NewAPIKeyService()

	expiry := time.Now().Add(24 * time.Hour)
	req := &CreateAPIKeyRequest{
		Name:      "LastUsed Test",
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}

	resp, err := service.CreateAPIKey("user-1", req)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	// First validation
	_, err = service.ValidateAPIKey(resp.Key)
	if err != nil {
		t.Fatalf("First validation failed: %v", err)
	}

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Second validation
	validatedKey, err := service.ValidateAPIKey(resp.Key)
	if err != nil {
		t.Fatalf("Second validation failed: %v", err)
	}

	if validatedKey.LastUsed == nil || validatedKey.LastUsed.IsZero() {
		t.Error("Expected LastUsed to be set")
	}
}

func TestAPIKeyMultipleKeys(t *testing.T) {
	service := NewAPIKeyService()

	expiry := time.Now().Add(24 * time.Hour)
	responses := make([]*CreateAPIKeyResponse, 5)

	// Create multiple keys
	for i := 0; i < 5; i++ {
		req := &CreateAPIKeyRequest{
			Name:      "Key-" + string(rune('A'+i)),
			Role:      RoleUser,
			ExpiresAt: &expiry,
		}
		resp, err := service.CreateAPIKey("user-"+string(rune('1'+i)), req)
		if err != nil {
			t.Fatalf("Failed to create key %d: %v", i, err)
		}
		responses[i] = resp
	}

	// Validate all keys
	for i, resp := range responses {
		validated, err := service.ValidateAPIKey(resp.Key)
		if err != nil {
			t.Errorf("Failed to validate key %d: %v", i, err)
		}
		if validated.ID != resp.ID {
			t.Errorf("Key %d: ID mismatch", i)
		}
	}

	// Revoke one key
	err := service.RevokeAPIKey(responses[2].ID)
	if err != nil {
		t.Fatalf("Failed to revoke key: %v", err)
	}

	// Verify revoked key is invalid but others are still valid
	for i, resp := range responses {
		_, err := service.ValidateAPIKey(resp.Key)
		if i == 2 {
			if err == nil {
				t.Error("Revoked key should be invalid")
			}
		} else {
			if err != nil {
				t.Errorf("Key %d should still be valid: %v", i, err)
			}
		}
	}
}

func TestAPIKeyListByUser(t *testing.T) {
	service := NewAPIKeyService()

	expiry := time.Now().Add(24 * time.Hour)
	userID := "user-list-test"

	// Create 3 keys for same user
	for i := 0; i < 3; i++ {
		req := &CreateAPIKeyRequest{
			Name:      "Key-" + string(rune('A'+i)),
			Role:      RoleUser,
			ExpiresAt: &expiry,
		}
		_, err := service.CreateAPIKey(userID, req)
		if err != nil {
			t.Fatalf("Failed to create key: %v", err)
		}
	}

	// Create key for different user
	req := &CreateAPIKeyRequest{
		Name:      "Other Key",
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}
	_, err := service.CreateAPIKey("other-user", req)
	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	// List keys for user
	keys := service.ListAPIKeys(userID)

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	for _, key := range keys {
		if key.UserID != userID {
			t.Errorf("Key belongs to wrong user: %s", key.UserID)
		}
	}
}

func TestAPIKeyUniqueness(t *testing.T) {
	service := NewAPIKeyService()

	expiry := time.Now().Add(24 * time.Hour)
	keys := make(map[string]bool)

	// Create 10 keys and verify they're all unique
	for i := 0; i < 10; i++ {
		req := &CreateAPIKeyRequest{
			Name:      "Key",
			Role:      RoleUser,
			ExpiresAt: &expiry,
		}
		resp, err := service.CreateAPIKey("user", req)
		if err != nil {
			t.Fatalf("Failed to create key: %v", err)
		}

		if keys[resp.Key] {
			t.Error("Generated duplicate key")
		}
		keys[resp.Key] = true

		if keys[resp.ID] {
			t.Error("Generated duplicate ID")
		}
		keys[resp.ID] = true
	}
}

func BenchmarkAPIKeyCreate(b *testing.B) {
	service := NewAPIKeyService()
	expiry := time.Now().Add(24 * time.Hour)
	req := &CreateAPIKeyRequest{
		Name:      "Bench Key",
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateAPIKey("user", req)
	}
}

func BenchmarkAPIKeyValidate(b *testing.B) {
	service := NewAPIKeyService()
	expiry := time.Now().Add(24 * time.Hour)
	req := &CreateAPIKeyRequest{
		Name:      "Bench Key",
		Role:      RoleUser,
		ExpiresAt: &expiry,
	}
	resp, _ := service.CreateAPIKey("user", req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ValidateAPIKey(resp.Key)
	}
}
