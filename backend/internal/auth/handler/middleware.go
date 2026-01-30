package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/service"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// UserContextKey is the context key for the authenticated user's claims.
	UserContextKey ContextKey = "auth_user"
	// UserIDContextKey is the context key for the user ID.
	UserIDContextKey ContextKey = "user_id"
	// TenantIDContextKey is the context key for the tenant ID.
	TenantIDContextKey ContextKey = "tenant_id"
	// RoleContextKey is the context key for the user's role.
	RoleContextKey ContextKey = "role"
)

// AuthMiddleware provides authentication middleware.
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth is middleware that requires a valid JWT token.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := extractBearerToken(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "token_invalid", "Authorization header missing or invalid")
			return
		}

		// Validate token
		claims, err := m.authService.ValidateToken(r.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrTokenExpired):
				writeError(w, http.StatusUnauthorized, "token_expired", "Token has expired")
			case errors.Is(err, domain.ErrTokenMalformed):
				writeError(w, http.StatusUnauthorized, "token_invalid", "Token is malformed")
			default:
				writeError(w, http.StatusUnauthorized, "token_invalid", "Token is invalid")
			}
			return
		}

		// Parse user ID from claims
		userID, err := claims.GetUserID()
		if err != nil {
			writeError(w, http.StatusUnauthorized, "token_invalid", "Invalid user ID in token")
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, claims)
		ctx = context.WithValue(ctx, UserIDContextKey, userID)
		ctx = context.WithValue(ctx, TenantIDContextKey, claims.TenantID)
		ctx = context.WithValue(ctx, RoleContextKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is middleware that requires a minimum role level.
func (m *AuthMiddleware) RequireRole(minRole domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleContextKey).(domain.Role)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
				return
			}

			// Check if user's role level meets the minimum requirement
			if role.Level() < minRole.Level() {
				writeError(w, http.StatusForbidden, "insufficient_role", "Insufficient role for this operation")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole is middleware that requires one of the specified roles.
func (m *AuthMiddleware) RequireAnyRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleContextKey).(domain.Role)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
				return
			}

			// Check if user has any of the required roles
			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeError(w, http.StatusForbidden, "insufficient_role", "Insufficient role for this operation")
		})
	}
}

// extractBearerToken extracts the JWT token from the Authorization header.
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return id, ok
}

// GetTenantID extracts the tenant ID from the request context.
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(TenantIDContextKey).(uuid.UUID)
	return id, ok
}

// GetRole extracts the role from the request context.
func GetRole(ctx context.Context) (domain.Role, bool) {
	role, ok := ctx.Value(RoleContextKey).(domain.Role)
	return role, ok
}

// GetClaims extracts the full claims from the request context.
func GetClaims(ctx context.Context) (*domain.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*domain.Claims)
	return claims, ok
}

// GetClientIP extracts the client IP address from the request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in format "IP:port", we need just the IP
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
