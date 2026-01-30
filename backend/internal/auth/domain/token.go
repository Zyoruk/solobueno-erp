package domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenPair represents an access and refresh token pair returned after login.
// This is a DTO, not persisted to the database.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"` // Always "Bearer"
	ExpiresIn    int       `json:"expires_in"` // Seconds until access token expires
	ExpiresAt    time.Time `json:"expires_at"` // Absolute expiration time
}

// NewTokenPair creates a new token pair with the given tokens and expiration.
func NewTokenPair(accessToken, refreshToken string, expiresAt time.Time) *TokenPair {
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(time.Until(expiresAt).Seconds()),
		ExpiresAt:    expiresAt,
	}
}

// Claims represents the JWT payload for access tokens.
type Claims struct {
	jwt.RegisteredClaims
	TenantID uuid.UUID `json:"tenant_id"`
	Role     Role      `json:"role"`
	Email    string    `json:"email"`
}

// NewClaims creates new JWT claims for a user session.
func NewClaims(userID, tenantID uuid.UUID, email string, role Role, expiresAt time.Time) *Claims {
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    "solobueno-erp",
			Audience:  jwt.ClaimStrings{"solobueno-api"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		TenantID: tenantID,
		Role:     role,
		Email:    email,
	}
}

// GetUserID parses and returns the user ID from the subject claim.
func (c *Claims) GetUserID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}

// IsExpired checks if the claims have expired.
func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}
