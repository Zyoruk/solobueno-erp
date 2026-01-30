// Package handler provides HTTP handlers for the auth module.
package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/service"
)

// --- Request DTOs ---

// LoginRequest is the request body for POST /login.
type LoginRequest struct {
	Email    string     `json:"email"`
	Password string     `json:"password"`
	TenantID *uuid.UUID `json:"tenant_id,omitempty"`
}

// RefreshRequest is the request body for POST /refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ChangePasswordRequest is the request body for POST /change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// PasswordResetRequest is the request body for POST /password-reset/request.
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetCompleteRequest is the request body for POST /password-reset/complete.
type PasswordResetCompleteRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// CreateUserRequest is the request body for POST /users.
type CreateUserRequest struct {
	Email     string      `json:"email"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Role      domain.Role `json:"role"`
}

// UpdateUserRequest is the request body for PATCH /users/{id}.
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// UpdateRoleRequest is the request body for PATCH /users/{id}/role.
type UpdateRoleRequest struct {
	Role domain.Role `json:"role"`
}

// --- Response DTOs ---

// LoginResponse is the response body for successful login.
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// TenantRequiredResponse is returned when user must select a tenant.
type TenantRequiredResponse struct {
	Error   ErrorDetail      `json:"error"`
}

// TenantOption represents a selectable tenant.
type TenantOption struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// TokenResponse is the response for token refresh.
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID                uuid.UUID `json:"id"`
	Email             string    `json:"email"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Role              string    `json:"role,omitempty"`
	TenantID          uuid.UUID `json:"tenant_id,omitempty"`
	IsActive          bool      `json:"is_active"`
	MustResetPassword bool      `json:"must_reset_password"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// MeResponse is the response for GET /me.
type MeResponse struct {
	ID                uuid.UUID          `json:"id"`
	Email             string             `json:"email"`
	FirstName         string             `json:"first_name"`
	LastName          string             `json:"last_name"`
	Role              string             `json:"role"`
	TenantID          uuid.UUID          `json:"tenant_id"`
	TenantName        string             `json:"tenant_name"`
	MustResetPassword bool               `json:"must_reset_password"`
	Tenants           []TenantRoleInfo   `json:"tenants"`
}

// TenantRoleInfo represents a user's role in a tenant.
type TenantRoleInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role string    `json:"role"`
}

// CreateUserResponse is the response for POST /users.
type CreateUserResponse struct {
	ID                uuid.UUID `json:"id"`
	Email             string    `json:"email"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Role              string    `json:"role"`
	TemporaryPassword string    `json:"temporary_password"`
	MustResetPassword bool      `json:"must_reset_password"`
	CreatedAt         time.Time `json:"created_at"`
}

// UserListResponse is the response for GET /users.
type UserListResponse struct {
	Data       []UserResponse `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

// Pagination contains pagination metadata.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// MessageResponse is a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// --- Error DTOs ---

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	RetryAfter int            `json:"retry_after,omitempty"`
	Tenants    []TenantOption `json:"tenants,omitempty"`
}

// --- Conversion Functions ---

// ToLoginResponse converts service response to API response.
func ToLoginResponse(resp *service.LoginResponse) *LoginResponse {
	return &LoginResponse{
		AccessToken:  resp.TokenPair.AccessToken,
		RefreshToken: resp.TokenPair.RefreshToken,
		TokenType:    resp.TokenPair.TokenType,
		ExpiresIn:    resp.TokenPair.ExpiresIn,
		ExpiresAt:    resp.TokenPair.ExpiresAt,
		User: UserResponse{
			ID:                resp.User.ID,
			Email:             resp.User.Email,
			FirstName:         resp.User.FirstName,
			LastName:          resp.User.LastName,
			Role:              string(resp.Role),
			TenantID:          resp.TenantID,
			IsActive:          resp.User.IsActive,
			MustResetPassword: resp.User.MustResetPwd,
			CreatedAt:         resp.User.CreatedAt,
		},
	}
}

// ToTokenResponse converts a token pair to API response.
func ToTokenResponse(tp *domain.TokenPair) *TokenResponse {
	return &TokenResponse{
		AccessToken:  tp.AccessToken,
		RefreshToken: tp.RefreshToken,
		TokenType:    tp.TokenType,
		ExpiresIn:    tp.ExpiresIn,
		ExpiresAt:    tp.ExpiresAt,
	}
}

// ToUserResponse converts a domain user to API response.
func ToUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:                user.ID,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		IsActive:          user.IsActive,
		MustResetPassword: user.MustResetPwd,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

// ToTenantOptions converts service tenant info to API format.
func ToTenantOptions(tenants []service.TenantInfo) []TenantOption {
	options := make([]TenantOption, len(tenants))
	for i, t := range tenants {
		options[i] = TenantOption{
			ID:   t.ID,
			Name: t.Name,
			Slug: t.Slug,
		}
	}
	return options
}
