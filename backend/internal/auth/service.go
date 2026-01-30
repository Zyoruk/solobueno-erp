// Package auth provides authentication and authorization functionality for Solobueno ERP.
//
// This package is the main entry point for the auth module and provides:
//   - User authentication (login, logout, token refresh)
//   - User management (CRUD operations)
//   - Role-based access control (RBAC)
//   - Password management (change, reset)
//   - Session management
//
// # Quick Start
//
// To use the auth module, create a Module instance and register its routes:
//
//	keyManager := jwt.NewKeyManager()
//	keyManager.LoadKeysFromEnv()
//
//	authModule, err := auth.NewModule(auth.ModuleConfig{
//	    DB:         db,
//	    KeyManager: keyManager,
//	    JWTConfig:  jwt.DefaultTokenGeneratorConfig(),
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	r := chi.NewRouter()
//	authModule.RegisterRoutes(r)
//
// # Endpoints
//
// The module exposes the following endpoints:
//
// Auth endpoints (base: /api/v1/auth):
//   - POST /login          - Authenticate and get tokens
//   - POST /refresh        - Refresh access token
//   - POST /logout         - Invalidate session
//   - GET  /me             - Get current user info
//   - POST /change-password - Change password
//   - POST /password-reset/request  - Request password reset
//   - POST /password-reset/complete - Complete password reset
//
// User endpoints (base: /api/v1/users):
//   - POST   /           - Create user (Manager+)
//   - GET    /           - List users (Manager+)
//   - GET    /{id}       - Get user (Manager+)
//   - PATCH  /{id}       - Update user (Manager+)
//   - PATCH  /{id}/role  - Change user role (Manager+)
//
// # Roles
//
// The module supports the following roles (highest to lowest):
//   - owner   (100) - Full access including billing
//   - admin   (90)  - Full access except billing
//   - manager (70)  - Can manage staff and operations
//   - cashier (50)  - Can process payments
//   - waiter  (40)  - Can take orders
//   - kitchen (30)  - Can view/update order status
//   - viewer  (10)  - Read-only access
//
// # Security
//
// The module implements several security measures:
//   - Argon2id password hashing with OWASP-recommended parameters
//   - RS256 JWT signing for access tokens
//   - Refresh token rotation on each use
//   - Rate limiting on login attempts (5/min/IP)
//   - All sessions invalidated on password change
//   - Audit logging for all auth events
package auth

import (
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/service"
)

// Re-export commonly used types for convenience

// Role represents a user's role within a tenant.
type Role = domain.Role

// Role constants
const (
	RoleOwner   = domain.RoleOwner
	RoleAdmin   = domain.RoleAdmin
	RoleManager = domain.RoleManager
	RoleCashier = domain.RoleCashier
	RoleWaiter  = domain.RoleWaiter
	RoleKitchen = domain.RoleKitchen
	RoleViewer  = domain.RoleViewer
)

// AuthService handles authentication operations.
type AuthService = service.AuthService

// UserService handles user management operations.
type UserService = service.UserService

// TokenService provides token generation and validation.
type TokenService = service.TokenService

// PasswordService handles password hashing and verification.
type PasswordService = service.PasswordService

// Error types
var (
	ErrInvalidCredentials = domain.ErrInvalidCredentials
	ErrAccountDisabled    = domain.ErrAccountDisabled
	ErrTokenExpired       = domain.ErrTokenExpired
	ErrTokenInvalid       = domain.ErrTokenInvalid
	ErrSessionRevoked     = domain.ErrSessionRevoked
	ErrInsufficientRole   = domain.ErrInsufficientRole
	ErrTenantRequired     = domain.ErrTenantRequired
	ErrEmailExists        = domain.ErrEmailExists
	ErrUserNotFound       = domain.ErrUserNotFound
	ErrRateLimitExceeded  = domain.ErrRateLimitExceeded
)
