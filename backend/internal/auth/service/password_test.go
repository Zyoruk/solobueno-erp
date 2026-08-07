package service

import (
	"strings"
	"testing"
)

func TestPasswordService_HashAndVerify(t *testing.T) {
	svc := NewPasswordService()

	password := "SecurePassword123"

	// Hash the password
	hash, err := svc.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	if hash == password {
		t.Error("hash should not equal plain password")
	}

	// Verify correct password
	match, err := svc.Verify(password, hash)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}

	if !match {
		t.Error("expected password to match")
	}

	// Verify wrong password
	match, err = svc.Verify("WrongPassword123", hash)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}

	if match {
		t.Error("expected wrong password to not match")
	}
}

func TestPasswordService_ValidatePassword(t *testing.T) {
	svc := NewPasswordService()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "SecurePass1",
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "Short1",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "no uppercase",
			password: "lowercase123",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "no lowercase",
			password: "UPPERCASE123",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "no digit",
			password: "NoDigitsHere",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "all lowercase",
			password: "alllowercase",
			wantErr:  ErrPasswordTooWeak,
		},
		{
			name:     "exactly 8 chars valid",
			password: "Valid1Aa",
			wantErr:  nil,
		},
		{
			name:     "complex valid password",
			password: "C0mpl3x!P@ssw0rd#",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestPasswordService_GenerateTemporaryPassword(t *testing.T) {
	svc := NewPasswordService()

	// Generate multiple passwords and check they're unique and valid
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		password, err := svc.GenerateTemporaryPassword()
		if err != nil {
			t.Fatalf("failed to generate temporary password: %v", err)
		}

		if len(password) != 12 {
			t.Errorf("expected password length 12, got %d", len(password))
		}

		if err := svc.ValidatePassword(password); err != nil {
			t.Errorf("generated password %q failed validation: %v", password, err)
		}

		if seen[password] {
			t.Error("generated duplicate password")
		}
		seen[password] = true
	}
}

func TestPasswordService_GenerateResetToken(t *testing.T) {
	svc := NewPasswordService()

	// Generate multiple tokens and check they're unique
	seenTokens := make(map[string]bool)
	seenHashes := make(map[string]bool)

	for i := 0; i < 10; i++ {
		plainToken, tokenHash, err := svc.GenerateResetToken()
		if err != nil {
			t.Fatalf("failed to generate reset token: %v", err)
		}

		if plainToken == "" {
			t.Error("expected non-empty plain token")
		}

		if tokenHash == "" {
			t.Error("expected non-empty token hash")
		}

		if plainToken == tokenHash {
			t.Error("plain token should not equal hash")
		}

		if seenTokens[plainToken] {
			t.Error("generated duplicate plain token")
		}
		seenTokens[plainToken] = true

		if seenHashes[tokenHash] {
			t.Error("generated duplicate token hash")
		}
		seenHashes[tokenHash] = true
	}
}

func TestPasswordService_HashResetToken(t *testing.T) {
	svc := NewPasswordService()

	plainToken, expectedHash, err := svc.GenerateResetToken()
	if err != nil {
		t.Fatalf("failed to generate reset token: %v", err)
	}

	// Hash the plain token and compare
	hash := svc.HashResetToken(plainToken)
	if hash != expectedHash {
		t.Errorf("HashResetToken returned different hash: got %s, want %s", hash, expectedHash)
	}

	// Different tokens should produce different hashes
	otherToken := plainToken + "modified"
	otherHash := svc.HashResetToken(otherToken)
	if otherHash == expectedHash {
		t.Error("different tokens should produce different hashes")
	}
}

func TestPasswordService_HashRefreshToken(t *testing.T) {
	svc := NewPasswordService()

	token := "test-refresh-token-12345"
	hash1 := svc.HashRefreshToken(token)
	hash2 := svc.HashRefreshToken(token)

	// Same token should produce same hash
	if hash1 != hash2 {
		t.Error("same token should produce same hash")
	}

	// Different tokens should produce different hashes
	differentHash := svc.HashRefreshToken(token + "different")
	if hash1 == differentHash {
		t.Error("different tokens should produce different hashes")
	}
}

func TestPasswordService_DifferentHashesForSamePassword(t *testing.T) {
	svc := NewPasswordService()

	password := "SamePassword123"

	hash1, err := svc.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	hash2, err := svc.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Argon2id uses random salt, so same password should produce different hashes
	if hash1 == hash2 {
		t.Error("expected different hashes for same password (due to random salt)")
	}

	// But both should verify correctly
	match1, _ := svc.Verify(password, hash1)
	match2, _ := svc.Verify(password, hash2)

	if !match1 || !match2 {
		t.Error("both hashes should verify correctly")
	}
}

func TestPasswordService_HashFormat(t *testing.T) {
	svc := NewPasswordService()

	password := "TestPassword123"
	hash, err := svc.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Argon2id hash format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("expected hash to start with $argon2id$, got %s", hash[:min(len(hash), 20)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
