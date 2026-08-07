package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrTokenMalformed is returned when the token is malformed.
	ErrTokenMalformed = errors.New("token is malformed")
	// ErrTokenExpired is returned when the token has expired.
	ErrTokenExpired = errors.New("token has expired")
	// ErrTokenInvalid is returned when the token is invalid.
	ErrTokenInvalid = errors.New("token is invalid")
	// ErrTokenSignatureInvalid is returned when the token signature is invalid.
	ErrTokenSignatureInvalid = errors.New("token signature is invalid")
)

// Claims represents the JWT payload for access tokens.
type Claims struct {
	jwt.RegisteredClaims
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
	Email    string    `json:"email"`
}

// TokenGenerator handles JWT token generation.
type TokenGenerator struct {
	keyManager       *KeyManager
	issuer           string
	audience         []string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

// TokenGeneratorConfig holds configuration for TokenGenerator.
type TokenGeneratorConfig struct {
	Issuer          string
	Audience        []string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// DefaultTokenGeneratorConfig returns default configuration.
func DefaultTokenGeneratorConfig() TokenGeneratorConfig {
	return TokenGeneratorConfig{
		Issuer:          "solobueno-erp",
		Audience:        []string{"solobueno-api"},
		AccessTokenTTL:  60 * time.Minute,   // 60 minutes per FR-003
		RefreshTokenTTL: 30 * 24 * time.Hour, // 30 days per FR-004
	}
}

// NewTokenGenerator creates a new TokenGenerator.
func NewTokenGenerator(keyManager *KeyManager, cfg TokenGeneratorConfig) *TokenGenerator {
	return &TokenGenerator{
		keyManager:      keyManager,
		issuer:          cfg.Issuer,
		audience:        cfg.Audience,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

// GenerateAccessToken generates a new JWT access token.
func (g *TokenGenerator) GenerateAccessToken(userID, tenantID uuid.UUID, email, role string) (string, time.Time, error) {
	privateKey, err := g.keyManager.GetPrivateKey()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get private key: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(g.accessTokenTTL)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    g.issuer,
			Audience:  g.audience,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		TenantID: tenantID,
		Role:     role,
		Email:    email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set key ID in header for key rotation support
	token.Header["kid"] = g.keyManager.GetKeyID()

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a new opaque refresh token.
// This is not a JWT - it's a random UUID that will be stored in the database.
func (g *TokenGenerator) GenerateRefreshToken() (string, time.Time) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(g.refreshTokenTTL)
	return token, expiresAt
}

// GetAccessTokenTTL returns the access token TTL.
func (g *TokenGenerator) GetAccessTokenTTL() time.Duration {
	return g.accessTokenTTL
}

// GetRefreshTokenTTL returns the refresh token TTL.
func (g *TokenGenerator) GetRefreshTokenTTL() time.Duration {
	return g.refreshTokenTTL
}

// TokenValidator handles JWT token validation.
type TokenValidator struct {
	keyManager *KeyManager
	issuer     string
	audience   []string
}

// NewTokenValidator creates a new TokenValidator.
func NewTokenValidator(keyManager *KeyManager, issuer string, audience []string) *TokenValidator {
	return &TokenValidator{
		keyManager: keyManager,
		issuer:     issuer,
		audience:   audience,
	}
}

// ValidateToken validates a JWT token and returns the claims.
func (v *TokenValidator) ValidateToken(tokenString string) (*Claims, error) {
	publicKey, err := v.keyManager.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, v.parseError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	// Validate issuer
	if claims.Issuer != v.issuer {
		return nil, ErrTokenInvalid
	}

	// Validate audience
	if !v.validateAudience(claims.Audience) {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// validateAudience checks if the token audience matches expected audience.
func (v *TokenValidator) validateAudience(tokenAudience []string) bool {
	for _, expected := range v.audience {
		for _, actual := range tokenAudience {
			if expected == actual {
				return true
			}
		}
	}
	return false
}

// parseError converts JWT library errors to our custom errors.
func (v *TokenValidator) parseError(err error) error {
	if errors.Is(err, jwt.ErrTokenMalformed) {
		return ErrTokenMalformed
	}
	if errors.Is(err, jwt.ErrTokenExpired) {
		return ErrTokenExpired
	}
	if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		return ErrTokenSignatureInvalid
	}
	return ErrTokenInvalid
}

// ParseUnverified parses a token without verifying the signature.
// Useful for extracting claims from expired tokens during refresh.
func ParseUnverified(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// GetUserIDFromClaims extracts the user ID from claims.
func GetUserIDFromClaims(claims *Claims) (uuid.UUID, error) {
	return uuid.Parse(claims.Subject)
}
