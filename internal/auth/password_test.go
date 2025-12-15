package auth

import (
	"errors"
	"testing"
)

func TestHashPassword(t *testing.T) {
	service := NewPasswordService()
	password := "SecurePassword123!"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}

	t.Logf("Hash length: %d", len(hash))
}

func TestVerifyPassword(t *testing.T) {
	service := NewPasswordService()
	password := "SecurePassword123!"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify correct password (password first, hash second)
	err = service.VerifyPassword(password, hash)
	if err != nil {
		t.Errorf("Failed to verify correct password: %v", err)
	}
}

func TestVerifyPasswordIncorrect(t *testing.T) {
	service := NewPasswordService()
	password := "CorrectPassword123!"
	wrongPassword := "WrongPassword456!"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify incorrect password
	err = service.VerifyPassword(wrongPassword, hash)
	if err == nil {
		t.Error("Expected error for incorrect password")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	service := NewPasswordService()
	password := "SamePassword123!"

	// Generate multiple hashes for same password
	hashes := make(map[string]bool)
	for i := 0; i < 5; i++ {
		hash, err := service.HashPassword(password)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		if hashes[hash] {
			t.Error("Generated duplicate hash for same password")
		}
		hashes[hash] = true
	}

	// All hashes should still verify the original password
	for hash := range hashes {
		err := service.VerifyPassword(password, hash)
		if err != nil {
			t.Errorf("Hash should verify original password: %v", err)
		}
	}
}

func TestHashPasswordVariousLengths(t *testing.T) {
	service := NewPasswordService()

	testCases := []struct {
		name      string
		password  string
		shouldErr bool
	}{
		{"short", "abc", true}, // Too short
		{"exactly8", "12345678", false},
		{"medium", "mediumPassword", false},
		{"long", "ThisIsAVeryLongPasswordThatShouldStillWorkFine123!", false},
		{"with_special", "P@$$w0rd!#%^&*()", false},
		{"unicode", "пароль密码🔐abcd", false},
		{"spaces", "password with spaces", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := service.HashPassword(tc.password)
			if tc.shouldErr {
				if err == nil {
					t.Error("Expected error for weak password")
				}
				return
			}
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			err = service.VerifyPassword(tc.password, hash)
			if err != nil {
				t.Errorf("Failed to verify password: %v", err)
			}
		})
	}
}

func TestVerifyPasswordEmptyHash(t *testing.T) {
	service := NewPasswordService()

	err := service.VerifyPassword("password", "")
	if err == nil {
		t.Error("Expected error for empty hash")
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	service := NewPasswordService()

	testCases := []struct {
		name string
		hash string
	}{
		{"random_string", "not-a-valid-bcrypt-hash"},
		{"base64", "dGhpcyBpcyBub3QgYSBoYXNo"},
		{"partial_bcrypt", "$2a$10$invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.VerifyPassword("password12", tc.hash)
			if err == nil {
				t.Error("Expected error for invalid hash")
			}
		})
	}
}

func TestHashPasswordCost(t *testing.T) {
	service := NewPasswordService()
	password := "TestPassword123!"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// bcrypt hash format: $2a$XX$... where XX is the cost
	// Should start with $2a$ or $2b$
	if len(hash) < 4 {
		t.Fatal("Hash too short")
	}

	if hash[0] != '$' || hash[1] != '2' {
		t.Error("Expected bcrypt hash format")
	}

	t.Logf("Hash prefix: %s", hash[:7])
}

func TestPasswordServiceConsistency(t *testing.T) {
	service1 := NewPasswordService()
	service2 := NewPasswordService()
	password := "ConsistencyTest123!"

	// Hash with service1
	hash, err := service1.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify with service2 (should work since bcrypt is standard)
	err = service2.VerifyPassword(password, hash)
	if err != nil {
		t.Error("Different service instance should verify hash")
	}
}

func TestPasswordTooShort(t *testing.T) {
	service := NewPasswordService()

	_, err := service.HashPassword("short")
	if err == nil {
		t.Error("Expected error for short password")
	}
	if !errors.Is(err, ErrPasswordTooWeak) {
		t.Errorf("Expected ErrPasswordTooWeak, got %v", err)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	service := NewPasswordService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.HashPassword("BenchmarkPassword123!")
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	service := NewPasswordService()
	hash, _ := service.HashPassword("BenchmarkPassword123!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.VerifyPassword("BenchmarkPassword123!", hash)
	}
}

func BenchmarkVerifyPasswordIncorrect(b *testing.B) {
	service := NewPasswordService()
	hash, _ := service.HashPassword("CorrectPassword123!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.VerifyPassword("WrongPassword456!", hash)
	}
}
