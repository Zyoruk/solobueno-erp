package domain

import "errors"

// Domain errors for the auth module.
// These are used by services and handlers to communicate specific error conditions.
var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrAccountNotFound    = errors.New("account not found")

	// Token errors
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrTokenMalformed     = errors.New("token is malformed")
	ErrSessionRevoked     = errors.New("session has been revoked")
	ErrSessionNotFound    = errors.New("session not found")
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")

	// Authorization errors
	ErrInsufficientRole   = errors.New("insufficient role for this operation")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")

	// Tenant errors
	ErrTenantRequired     = errors.New("tenant selection required")
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrTenantInactive     = errors.New("tenant is inactive")
	ErrUserNotInTenant    = errors.New("user does not belong to this tenant")

	// User management errors
	ErrEmailExists        = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrCannotManageRole   = errors.New("cannot manage users with this role")
	ErrCannotAssignRole   = errors.New("cannot assign this role")

	// Password errors
	ErrPasswordWeak         = errors.New("password does not meet requirements")
	ErrPasswordIncorrect    = errors.New("current password is incorrect")
	ErrPasswordResetExpired = errors.New("password reset token has expired")
	ErrPasswordResetUsed    = errors.New("password reset token has already been used")
	ErrPasswordResetInvalid = errors.New("password reset token is invalid")

	// Rate limiting errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// AuthError wraps an error with additional context.
type AuthError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// NewAuthError creates a new AuthError.
func NewAuthError(code, message string, err error) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsAuthError checks if the error is an AuthError with the specified code.
func IsAuthError(err error, code string) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == code
	}
	return false
}
