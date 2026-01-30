package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/pkg/jwt"
)

// TokenService provides token generation and validation using domain types.
type TokenService struct {
	generator       *jwt.TokenGenerator
	validator       *jwt.TokenValidator
	passwordService *PasswordService
}

// NewTokenService creates a new TokenService.
func NewTokenService(keyManager *jwt.KeyManager, cfg jwt.TokenGeneratorConfig) *TokenService {
	return &TokenService{
		generator:       jwt.NewTokenGenerator(keyManager, cfg),
		validator:       jwt.NewTokenValidator(keyManager, cfg.Issuer, cfg.Audience),
		passwordService: NewPasswordService(),
	}
}

// GenerateTokenPair generates a new access and refresh token pair.
func (s *TokenService) GenerateTokenPair(user *domain.User, tenantID uuid.UUID, role domain.Role) (*domain.TokenPair, string, error) {
	// Generate access token
	accessToken, expiresAt, err := s.generator.GenerateAccessToken(user.ID, tenantID, user.Email, string(role))
	if err != nil {
		return nil, "", err
	}

	// Generate refresh token (opaque)
	refreshToken, _ := s.generator.GenerateRefreshToken()

	// Hash the refresh token for storage
	refreshTokenHash := s.passwordService.HashRefreshToken(refreshToken)

	tokenPair := domain.NewTokenPair(accessToken, refreshToken, expiresAt)

	return tokenPair, refreshTokenHash, nil
}

// ValidateAccessToken validates an access token and returns the claims.
func (s *TokenService) ValidateAccessToken(tokenString string) (*domain.Claims, error) {
	jwtClaims, err := s.validator.ValidateToken(tokenString)
	if err != nil {
		switch err {
		case jwt.ErrTokenExpired:
			return nil, domain.ErrTokenExpired
		case jwt.ErrTokenMalformed:
			return nil, domain.ErrTokenMalformed
		case jwt.ErrTokenSignatureInvalid:
			return nil, domain.ErrTokenInvalid
		default:
			return nil, domain.ErrTokenInvalid
		}
	}

	// Convert JWT claims to domain claims
	claims := &domain.Claims{
		RegisteredClaims: jwtClaims.RegisteredClaims,
		TenantID:         jwtClaims.TenantID,
		Role:             domain.Role(jwtClaims.Role),
		Email:            jwtClaims.Email,
	}

	return claims, nil
}

// HashRefreshToken hashes a plain refresh token for comparison.
func (s *TokenService) HashRefreshToken(plainToken string) string {
	return s.passwordService.HashRefreshToken(plainToken)
}

// GetAccessTokenTTL returns the access token TTL.
func (s *TokenService) GetAccessTokenTTL() time.Duration {
	return s.generator.GetAccessTokenTTL()
}

// GetRefreshTokenTTL returns the refresh token TTL.
func (s *TokenService) GetRefreshTokenTTL() time.Duration {
	return s.generator.GetRefreshTokenTTL()
}

// GetRefreshTokenExpiry returns the expiry time for a new refresh token.
func (s *TokenService) GetRefreshTokenExpiry() time.Time {
	return time.Now().Add(s.generator.GetRefreshTokenTTL())
}
