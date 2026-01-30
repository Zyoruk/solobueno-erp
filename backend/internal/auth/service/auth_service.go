package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository"
)

// AuthService handles authentication operations.
type AuthService struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	eventRepo    repository.AuthEventRepository
	tenantRepo   repository.TenantRepository
	roleRepo     repository.UserTenantRoleRepository
	tokenService *TokenService
	passwordSvc  *PasswordService
	rateLimiter  RateLimiter
}

// AuthServiceConfig holds configuration for AuthService.
type AuthServiceConfig struct {
	UserRepo     repository.UserRepository
	SessionRepo  repository.SessionRepository
	EventRepo    repository.AuthEventRepository
	TenantRepo   repository.TenantRepository
	RoleRepo     repository.UserTenantRoleRepository
	TokenService *TokenService
	RateLimiter  RateLimiter
}

// NewAuthService creates a new AuthService.
func NewAuthService(cfg AuthServiceConfig) *AuthService {
	return &AuthService{
		userRepo:     cfg.UserRepo,
		sessionRepo:  cfg.SessionRepo,
		eventRepo:    cfg.EventRepo,
		tenantRepo:   cfg.TenantRepo,
		roleRepo:     cfg.RoleRepo,
		tokenService: cfg.TokenService,
		passwordSvc:  NewPasswordService(),
		rateLimiter:  cfg.RateLimiter,
	}
}

// LoginRequest contains the data needed for login.
type LoginRequest struct {
	Email     string
	Password  string
	TenantID  *uuid.UUID // Optional, required if user belongs to multiple tenants
	IPAddress string
	UserAgent string
}

// LoginResponse contains the login result.
type LoginResponse struct {
	TokenPair *domain.TokenPair
	User      *domain.User
	TenantID  uuid.UUID
	Role      domain.Role
	// Tenants is set when user needs to select a tenant
	Tenants []TenantInfo
}

// TenantInfo contains basic tenant information.
type TenantInfo struct {
	ID   uuid.UUID   `json:"id"`
	Name string      `json:"name"`
	Slug string      `json:"slug"`
	Role domain.Role `json:"role"`
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Check rate limit
	if s.rateLimiter != nil {
		allowed, err := s.rateLimiter.Allow(ctx, req.IPAddress)
		if err != nil {
			return nil, err
		}
		if !allowed {
			s.logEvent(ctx, domain.EventLoginFailed, nil, nil, req.IPAddress, req.UserAgent, map[string]interface{}{
				"email":  req.Email,
				"reason": "rate_limit_exceeded",
			})
			return nil, domain.ErrRateLimitExceeded
		}
	}

	// Find user by email with tenant roles
	user, err := s.userRepo.FindByEmailWithTenants(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logEvent(ctx, domain.EventLoginFailed, nil, nil, req.IPAddress, req.UserAgent, map[string]interface{}{
				"email":  req.Email,
				"reason": "user_not_found",
			})
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	match, err := s.passwordSvc.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !match {
		s.logEvent(ctx, domain.EventLoginFailed, &user.ID, nil, req.IPAddress, req.UserAgent, map[string]interface{}{
			"reason": "invalid_password",
		})
		return nil, domain.ErrInvalidCredentials
	}

	// Check if account is active
	if !user.CanLogin() {
		s.logEvent(ctx, domain.EventLoginFailed, &user.ID, nil, req.IPAddress, req.UserAgent, map[string]interface{}{
			"reason": "account_disabled",
		})
		return nil, domain.ErrAccountDisabled
	}

	// Handle tenant selection
	var selectedTenantID uuid.UUID
	var selectedRole domain.Role

	if len(user.TenantRoles) == 0 {
		return nil, domain.ErrUserNotInTenant
	}

	if len(user.TenantRoles) == 1 {
		// Single tenant - auto-select
		selectedTenantID = user.TenantRoles[0].TenantID
		selectedRole = user.TenantRoles[0].Role
	} else if req.TenantID != nil {
		// Multiple tenants - use provided tenant ID
		found := false
		for _, tr := range user.TenantRoles {
			if tr.TenantID == *req.TenantID {
				selectedTenantID = tr.TenantID
				selectedRole = tr.Role
				found = true
				break
			}
		}
		if !found {
			return nil, domain.ErrUserNotInTenant
		}
	} else {
		// Multiple tenants - require selection
		tenants := make([]TenantInfo, 0, len(user.TenantRoles))
		for _, tr := range user.TenantRoles {
			tenants = append(tenants, TenantInfo{
				ID:   tr.TenantID,
				Name: tr.Tenant.Name,
				Slug: tr.Tenant.Slug,
				Role: tr.Role,
			})
		}
		return &LoginResponse{
			User:    user,
			Tenants: tenants,
		}, domain.ErrTenantRequired
	}

	// Verify tenant is active
	tenant, err := s.tenantRepo.FindByID(ctx, selectedTenantID)
	if err != nil {
		return nil, err
	}
	if !tenant.IsOperational() {
		return nil, domain.ErrTenantInactive
	}

	// Generate tokens
	tokenPair, refreshTokenHash, err := s.tokenService.GenerateTokenPair(user, selectedTenantID, selectedRole)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		TenantID:     selectedTenantID,
		RefreshToken: refreshTokenHash,
		DeviceInfo:   req.UserAgent,
		IPAddress:    req.IPAddress,
		ExpiresAt:    s.tokenService.GetRefreshTokenExpiry(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Log successful login
	s.logEvent(ctx, domain.EventLoginSuccess, &user.ID, &selectedTenantID, req.IPAddress, req.UserAgent, nil)

	return &LoginResponse{
		TokenPair: tokenPair,
		User:      user,
		TenantID:  selectedTenantID,
		Role:      selectedRole,
	}, nil
}

// RefreshRequest contains the data needed for token refresh.
type RefreshRequest struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

// Refresh refreshes an access token using a refresh token.
func (s *AuthService) Refresh(ctx context.Context, req RefreshRequest) (*domain.TokenPair, error) {
	// Hash the provided refresh token
	tokenHash := s.tokenService.HashRefreshToken(req.RefreshToken)

	// Find session by refresh token
	session, err := s.sessionRepo.FindByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil, domain.ErrRefreshTokenInvalid
		}
		return nil, err
	}

	// Check if session is valid
	if !session.IsValid() {
		if session.IsRevoked() {
			return nil, domain.ErrSessionRevoked
		}
		return nil, domain.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Check if user is still active
	if !user.CanLogin() {
		return nil, domain.ErrAccountDisabled
	}

	// Get user's role in the tenant
	role := user.GetRoleForTenant(session.TenantID)
	if role == "" {
		// User no longer belongs to this tenant
		return nil, domain.ErrUserNotInTenant
	}

	// Generate new token pair
	tokenPair, newRefreshTokenHash, err := s.tokenService.GenerateTokenPair(user, session.TenantID, role)
	if err != nil {
		return nil, err
	}

	// Revoke old session
	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return nil, err
	}

	// Create new session with rotated refresh token
	newSession := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		TenantID:     session.TenantID,
		RefreshToken: newRefreshTokenHash,
		DeviceInfo:   req.UserAgent,
		IPAddress:    req.IPAddress,
		ExpiresAt:    s.tokenService.GetRefreshTokenExpiry(),
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	// Log token refresh
	s.logEvent(ctx, domain.EventTokenRefresh, &user.ID, &session.TenantID, req.IPAddress, req.UserAgent, nil)

	return tokenPair, nil
}

// Logout invalidates a user's session.
func (s *AuthService) Logout(ctx context.Context, refreshToken, ipAddress string) error {
	// Hash the provided refresh token
	tokenHash := s.tokenService.HashRefreshToken(refreshToken)

	// Find session
	session, err := s.sessionRepo.FindByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			// Already logged out or invalid token - treat as success
			return nil
		}
		return err
	}

	// Revoke session
	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return err
	}

	// Log logout
	s.logEvent(ctx, domain.EventLogout, &session.UserID, &session.TenantID, ipAddress, "", nil)

	return nil
}

// LogoutAll invalidates all sessions for a user.
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID, ipAddress string) error {
	if err := s.sessionRepo.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}

	s.logEvent(ctx, domain.EventSessionRevoked, &userID, nil, ipAddress, "", map[string]interface{}{
		"scope": "all_sessions",
	})

	return nil
}

// ValidateToken validates an access token and returns the claims.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*domain.Claims, error) {
	return s.tokenService.ValidateAccessToken(token)
}

// logEvent logs an authentication event.
func (s *AuthService) logEvent(ctx context.Context, eventType domain.AuthEventType, userID, tenantID *uuid.UUID, ipAddress, userAgent string, metadata map[string]interface{}) {
	event := domain.NewAuthEvent(eventType, userID, tenantID, ipAddress, userAgent)
	if metadata != nil {
		event.Metadata = metadata
	}
	// Fire and forget - don't fail login if logging fails
	_ = s.eventRepo.Create(ctx, event)
}
