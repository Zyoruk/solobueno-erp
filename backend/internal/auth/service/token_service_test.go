package service

import (
	"testing"
)

func TestTokenService_HashRefreshToken(t *testing.T) {
	// Create a minimal token service just to test hashing
	svc := &TokenService{
		passwordService: NewPasswordService(),
	}

	token := "test-refresh-token"
	hash1 := svc.HashRefreshToken(token)
	hash2 := svc.HashRefreshToken(token)

	// Same token should produce same hash
	if hash1 != hash2 {
		t.Error("Same token should produce same hash")
	}

	// Different tokens should produce different hashes
	hash3 := svc.HashRefreshToken("different-token")
	if hash1 == hash3 {
		t.Error("Different tokens should produce different hashes")
	}

	// Hash should not be empty
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}

	// Hash should be different from original token
	if hash1 == token {
		t.Error("Hash should not equal original token")
	}
}
