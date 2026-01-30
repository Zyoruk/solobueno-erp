package domain

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestNewTokenPair(t *testing.T) {
	accessToken := "access.token.here"
	refreshToken := "refresh-token-uuid"
	expiresAt := time.Now().Add(time.Hour)

	tp := NewTokenPair(accessToken, refreshToken, expiresAt)

	if tp.AccessToken != accessToken {
		t.Error("AccessToken mismatch")
	}
	if tp.RefreshToken != refreshToken {
		t.Error("RefreshToken mismatch")
	}
	if tp.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", tp.TokenType, "Bearer")
	}
	if tp.ExpiresIn < 3500 || tp.ExpiresIn > 3600 {
		t.Errorf("ExpiresIn = %d, expected ~3600", tp.ExpiresIn)
	}
	if !tp.ExpiresAt.Equal(expiresAt) {
		t.Error("ExpiresAt mismatch")
	}
}

func TestNewClaims(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	email := "test@example.com"
	role := RoleManager
	expiresAt := time.Now().Add(time.Hour)

	claims := NewClaims(userID, tenantID, email, role, expiresAt)

	if claims.Subject != userID.String() {
		t.Error("Subject mismatch")
	}
	if claims.TenantID != tenantID {
		t.Error("TenantID mismatch")
	}
	if claims.Email != email {
		t.Error("Email mismatch")
	}
	if claims.Role != role {
		t.Error("Role mismatch")
	}
	if claims.Issuer != "solobueno-erp" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "solobueno-erp")
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "solobueno-api" {
		t.Error("Audience mismatch")
	}
	if claims.ID == "" {
		t.Error("ID should be set")
	}
}

func TestClaims_GetUserID(t *testing.T) {
	userID := uuid.New()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID.String(),
		},
	}

	got, err := claims.GetUserID()
	if err != nil {
		t.Errorf("GetUserID() error: %v", err)
	}
	if got != userID {
		t.Errorf("GetUserID() = %v, want %v", got, userID)
	}

	// Test invalid UUID
	claims.Subject = "invalid-uuid"
	_, err = claims.GetUserID()
	if err == nil {
		t.Error("GetUserID() should error on invalid UUID")
	}
}

func TestClaims_IsExpired(t *testing.T) {
	// Not expired
	claims1 := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	if claims1.IsExpired() {
		t.Error("Future expiry should not be expired")
	}

	// Expired
	claims2 := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	if !claims2.IsExpired() {
		t.Error("Past expiry should be expired")
	}

	// No expiry set
	claims3 := &Claims{}
	if !claims3.IsExpired() {
		t.Error("No expiry should be considered expired")
	}
}
