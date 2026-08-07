package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"
)

// generateTestKeyPair generates a test RSA key pair.
func generateTestKeyPair(t *testing.T) *KeyManager {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	km := NewKeyManager()
	km.privateKey = privateKey
	km.publicKey = &privateKey.PublicKey
	km.keyID = "test-key-1"

	return km
}

func TestTokenGeneratorAndValidator(t *testing.T) {
	km := generateTestKeyPair(t)

	cfg := DefaultTokenGeneratorConfig()
	generator := NewTokenGenerator(km, cfg)
	validator := NewTokenValidator(km, cfg.Issuer, cfg.Audience)

	userID := uuid.New()
	tenantID := uuid.New()
	email := "test@example.com"
	role := "manager"

	// Generate access token
	token, expiresAt, err := generator.GenerateAccessToken(userID, tenantID, email, role)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	if expiresAt.Before(time.Now()) {
		t.Error("expected expiration in the future")
	}

	// Validate token
	claims, err := validator.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.Subject != userID.String() {
		t.Errorf("expected subject %s, got %s", userID.String(), claims.Subject)
	}

	if claims.TenantID != tenantID {
		t.Errorf("expected tenant ID %s, got %s", tenantID, claims.TenantID)
	}

	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}

	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}
}

func TestRefreshTokenGeneration(t *testing.T) {
	km := generateTestKeyPair(t)
	cfg := DefaultTokenGeneratorConfig()
	generator := NewTokenGenerator(km, cfg)

	token1, expiry1 := generator.GenerateRefreshToken()
	token2, expiry2 := generator.GenerateRefreshToken()

	if token1 == "" || token2 == "" {
		t.Error("expected non-empty refresh tokens")
	}

	if token1 == token2 {
		t.Error("expected unique refresh tokens")
	}

	expectedExpiry := time.Now().Add(cfg.RefreshTokenTTL)
	if expiry1.Before(expectedExpiry.Add(-time.Minute)) || expiry1.After(expectedExpiry.Add(time.Minute)) {
		t.Error("unexpected refresh token expiry")
	}

	if expiry2.Before(expectedExpiry.Add(-time.Minute)) || expiry2.After(expectedExpiry.Add(time.Minute)) {
		t.Error("unexpected refresh token expiry")
	}
}

func TestInvalidToken(t *testing.T) {
	km := generateTestKeyPair(t)
	cfg := DefaultTokenGeneratorConfig()
	validator := NewTokenValidator(km, cfg.Issuer, cfg.Audience)

	tests := []struct {
		name  string
		token string
		err   error
	}{
		{
			name:  "empty token",
			token: "",
			err:   ErrTokenMalformed,
		},
		{
			name:  "malformed token",
			token: "not.a.valid.token",
			err:   ErrTokenMalformed,
		},
		{
			name:  "invalid signature",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.invalid",
			err:   ErrTokenInvalid, // Parser will fail before signature check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ValidateToken(tt.token)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestDifferentKeyPairs(t *testing.T) {
	// Generate token with one key pair
	km1 := generateTestKeyPair(t)
	cfg := DefaultTokenGeneratorConfig()
	generator := NewTokenGenerator(km1, cfg)

	token, _, err := generator.GenerateAccessToken(uuid.New(), uuid.New(), "test@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Try to validate with different key pair
	km2 := generateTestKeyPair(t)
	validator := NewTokenValidator(km2, cfg.Issuer, cfg.Audience)

	_, err = validator.ValidateToken(token)
	if err == nil {
		t.Error("expected error when validating with different key")
	}
}

func TestParseUnverified(t *testing.T) {
	km := generateTestKeyPair(t)
	cfg := DefaultTokenGeneratorConfig()
	generator := NewTokenGenerator(km, cfg)

	userID := uuid.New()
	tenantID := uuid.New()
	email := "test@example.com"
	role := "manager"

	token, _, err := generator.GenerateAccessToken(userID, tenantID, email, role)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Parse without verification
	claims, err := ParseUnverified(token)
	if err != nil {
		t.Fatalf("failed to parse unverified token: %v", err)
	}

	if claims.Subject != userID.String() {
		t.Errorf("expected subject %s, got %s", userID.String(), claims.Subject)
	}

	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
}

func TestGetUserIDFromClaims(t *testing.T) {
	km := generateTestKeyPair(t)
	cfg := DefaultTokenGeneratorConfig()
	generator := NewTokenGenerator(km, cfg)
	validator := NewTokenValidator(km, cfg.Issuer, cfg.Audience)

	userID := uuid.New()
	token, _, err := generator.GenerateAccessToken(userID, uuid.New(), "test@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := validator.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	extractedID, err := GetUserIDFromClaims(claims)
	if err != nil {
		t.Fatalf("failed to extract user ID: %v", err)
	}

	if extractedID != userID {
		t.Errorf("expected user ID %s, got %s", userID, extractedID)
	}
}

func TestKeyManagerNoKey(t *testing.T) {
	km := NewKeyManager()

	_, err := km.GetPrivateKey()
	if err != ErrKeyNotLoaded {
		t.Errorf("expected ErrKeyNotLoaded, got %v", err)
	}

	_, err = km.GetPublicKey()
	if err != ErrKeyNotLoaded {
		t.Errorf("expected ErrKeyNotLoaded, got %v", err)
	}

	if km.HasPrivateKey() {
		t.Error("expected HasPrivateKey to be false")
	}

	if km.HasPublicKey() {
		t.Error("expected HasPublicKey to be false")
	}
}

func TestTokenTTL(t *testing.T) {
	km := generateTestKeyPair(t)
	cfg := DefaultTokenGeneratorConfig()
	generator := NewTokenGenerator(km, cfg)

	if generator.GetAccessTokenTTL() != cfg.AccessTokenTTL {
		t.Errorf("expected access TTL %v, got %v", cfg.AccessTokenTTL, generator.GetAccessTokenTTL())
	}

	if generator.GetRefreshTokenTTL() != cfg.RefreshTokenTTL {
		t.Errorf("expected refresh TTL %v, got %v", cfg.RefreshTokenTTL, generator.GetRefreshTokenTTL())
	}
}
