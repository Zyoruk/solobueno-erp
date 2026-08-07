package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository"
	"github.com/solobueno/erp/pkg/jwt"
)

// testKeyPair generates RSA keys for testing.
func testKeyPair(t *testing.T) (*rsa.PrivateKey, []byte, []byte) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return privateKey, privateKeyPEM, publicKeyPEM
}

func setupAuthService(t *testing.T) (*AuthService, *repository.MockUserRepository, *repository.MockSessionRepository, *repository.MockTenantRepository, *repository.MockAuthEventRepository) {
	t.Helper()

	_, privatePEM, publicPEM := testKeyPair(t)

	km := jwt.NewKeyManager()
	if err := km.LoadPrivateKeyFromPEM(privatePEM); err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}
	if err := km.LoadPublicKeyFromPEM(publicPEM); err != nil {
		t.Fatalf("failed to load public key: %v", err)
	}

	tokenSvc := NewTokenService(km, jwt.DefaultTokenGeneratorConfig())

	userRepo := repository.NewMockUserRepository()
	sessionRepo := repository.NewMockSessionRepository()
	eventRepo := repository.NewMockAuthEventRepository()
	tenantRepo := repository.NewMockTenantRepository()
	roleRepo := repository.NewMockUserTenantRoleRepository()

	authSvc := NewAuthService(AuthServiceConfig{
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		EventRepo:    eventRepo,
		TenantRepo:   tenantRepo,
		RoleRepo:     roleRepo,
		TokenService: tokenSvc,
		RateLimiter:  nil, // No rate limiting for tests
	})

	return authSvc, userRepo, sessionRepo, tenantRepo, eventRepo
}

func TestAuthService_Login_Success(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, eventRepo := setupAuthService(t)
	ctx := context.Background()

	// Create test user
	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		FirstName:    "Test",
		LastName:     "User",
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{
				ID:       uuid.New(),
				UserID:   userID,
				TenantID: tenantID,
				Role:     domain.RoleManager,
			},
		},
	}
	userRepo.AddUser(user)

	tenant := &domain.Tenant{
		ID:       tenantID,
		Name:     "Test Restaurant",
		Slug:     "test-restaurant",
		IsActive: true,
	}
	tenantRepo.AddTenant(tenant)

	// Test login
	resp, err := authSvc.Login(ctx, LoginRequest{
		Email:     "test@example.com",
		Password:  "Password123!",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	})

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.TokenPair == nil {
		t.Fatal("TokenPair should not be nil")
	}
	if resp.TokenPair.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if resp.TokenPair.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}
	if resp.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", resp.TenantID, tenantID)
	}
	if resp.Role != domain.RoleManager {
		t.Errorf("Role = %v, want %v", resp.Role, domain.RoleManager)
	}

	// Check event was logged
	events := eventRepo.GetEvents()
	if len(events) == 0 {
		t.Error("Should have logged login event")
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("CorrectPassword123!")

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleManager},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, IsActive: true})

	_, err := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	})

	if err != domain.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authSvc, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	_, err := authSvc.Login(ctx, LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "Password123!",
	})

	if err != domain.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_AccountDisabled(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     false, // Disabled
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleManager},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, IsActive: true})

	_, err := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})

	if err != domain.ErrAccountDisabled {
		t.Errorf("Expected ErrAccountDisabled, got %v", err)
	}
}

func TestAuthService_Login_TenantRequired(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenant1ID := uuid.New()
	tenant2ID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	tenant1 := &domain.Tenant{ID: tenant1ID, Name: "Restaurant 1", Slug: "r1", IsActive: true}
	tenant2 := &domain.Tenant{ID: tenant2ID, Name: "Restaurant 2", Slug: "r2", IsActive: true}

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenant1ID, Role: domain.RoleManager, Tenant: *tenant1},
			{TenantID: tenant2ID, Role: domain.RoleWaiter, Tenant: *tenant2},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(tenant1)
	tenantRepo.AddTenant(tenant2)

	resp, err := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})

	if err != domain.ErrTenantRequired {
		t.Errorf("Expected ErrTenantRequired, got %v", err)
	}

	if resp == nil || len(resp.Tenants) != 2 {
		t.Error("Should return list of tenants")
	}
}

func TestAuthService_Login_WithTenantSelection(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenant1ID := uuid.New()
	tenant2ID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	tenant1 := &domain.Tenant{ID: tenant1ID, Name: "Restaurant 1", Slug: "r1", IsActive: true}
	tenant2 := &domain.Tenant{ID: tenant2ID, Name: "Restaurant 2", Slug: "r2", IsActive: true}

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenant1ID, Role: domain.RoleManager, Tenant: *tenant1},
			{TenantID: tenant2ID, Role: domain.RoleWaiter, Tenant: *tenant2},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(tenant1)
	tenantRepo.AddTenant(tenant2)

	// Login with tenant selection
	resp, err := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
		TenantID: &tenant2ID,
	})

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.TenantID != tenant2ID {
		t.Errorf("TenantID = %v, want %v", resp.TenantID, tenant2ID)
	}
	if resp.Role != domain.RoleWaiter {
		t.Errorf("Role = %v, want %v", resp.Role, domain.RoleWaiter)
	}
}

func TestAuthService_Login_RateLimited(t *testing.T) {
	_, privatePEM, publicPEM := testKeyPair(t)

	km := jwt.NewKeyManager()
	km.LoadPrivateKeyFromPEM(privatePEM)
	km.LoadPublicKeyFromPEM(publicPEM)

	tokenSvc := NewTokenService(km, jwt.DefaultTokenGeneratorConfig())

	// Create rate limiter that denies everything
	rateLimiter := NewMemoryRateLimiter(RateLimiterConfig{
		MaxRequests: 0, // Deny all
		Window:      time.Minute,
		KeyPrefix:   "test:",
	})

	authSvc := NewAuthService(AuthServiceConfig{
		UserRepo:     repository.NewMockUserRepository(),
		SessionRepo:  repository.NewMockSessionRepository(),
		EventRepo:    repository.NewMockAuthEventRepository(),
		TenantRepo:   repository.NewMockTenantRepository(),
		RoleRepo:     repository.NewMockUserTenantRoleRepository(),
		TokenService: tokenSvc,
		RateLimiter:  rateLimiter,
	})

	ctx := context.Background()
	_, err := authSvc.Login(ctx, LoginRequest{
		Email:     "test@example.com",
		Password:  "Password123!",
		IPAddress: "127.0.0.1",
	})

	if err != domain.ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got %v", err)
	}
}

func TestAuthService_Refresh_Success(t *testing.T) {
	authSvc, userRepo, sessionRepo, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleManager},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, IsActive: true})

	// First, login to get tokens
	loginResp, err := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Now refresh
	tokenPair, err := authSvc.Refresh(ctx, RefreshRequest{
		RefreshToken: loginResp.TokenPair.RefreshToken,
		IPAddress:    "127.0.0.1",
		UserAgent:    "TestAgent",
	})

	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("New AccessToken should not be empty")
	}
	if tokenPair.RefreshToken == "" {
		t.Error("New RefreshToken should not be empty")
	}
	if tokenPair.RefreshToken == loginResp.TokenPair.RefreshToken {
		t.Error("RefreshToken should be rotated")
	}

	// Old session should be revoked
	_ = sessionRepo // Session is managed internally
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	authSvc, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	_, err := authSvc.Refresh(ctx, RefreshRequest{
		RefreshToken: "invalid-token",
	})

	if err != domain.ErrRefreshTokenInvalid {
		t.Errorf("Expected ErrRefreshTokenInvalid, got %v", err)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleManager},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, IsActive: true})

	// Login
	loginResp, _ := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})

	// Logout
	err := authSvc.Logout(ctx, loginResp.TokenPair.RefreshToken, "127.0.0.1")
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Trying to refresh should fail
	_, err = authSvc.Refresh(ctx, RefreshRequest{
		RefreshToken: loginResp.TokenPair.RefreshToken,
	})

	if err != domain.ErrSessionRevoked && err != domain.ErrRefreshTokenInvalid {
		t.Errorf("Expected session error, got %v", err)
	}
}

func TestAuthService_Logout_InvalidToken(t *testing.T) {
	authSvc, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Logout with invalid token should succeed (idempotent)
	err := authSvc.Logout(ctx, "invalid-token", "127.0.0.1")
	if err != nil {
		t.Errorf("Logout with invalid token should succeed, got %v", err)
	}
}

func TestAuthService_LogoutAll(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleManager},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, IsActive: true})

	// Login twice
	loginResp1, _ := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})
	loginResp2, _ := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})

	// Logout all
	err := authSvc.LogoutAll(ctx, userID, "127.0.0.1")
	if err != nil {
		t.Fatalf("LogoutAll failed: %v", err)
	}

	// Both sessions should be invalid
	_, err = authSvc.Refresh(ctx, RefreshRequest{RefreshToken: loginResp1.TokenPair.RefreshToken})
	if err == nil {
		t.Error("First session should be revoked")
	}

	_, err = authSvc.Refresh(ctx, RefreshRequest{RefreshToken: loginResp2.TokenPair.RefreshToken})
	if err == nil {
		t.Error("Second session should be revoked")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("Password123!")

	user := &domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleManager},
		},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, IsActive: true})

	// Login
	loginResp, _ := authSvc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})

	// Validate token
	claims, err := authSvc.ValidateToken(ctx, loginResp.TokenPair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", claims.Email, "test@example.com")
	}
	if claims.Role != domain.RoleManager {
		t.Errorf("Role = %v, want %v", claims.Role, domain.RoleManager)
	}
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	authSvc, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	_, err := authSvc.ValidateToken(ctx, "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
