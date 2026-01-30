package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles POST /login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Email and password are required")
		return
	}

	loginReq := service.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		TenantID:  req.TenantID,
		IPAddress: GetClientIP(r),
		UserAgent: r.UserAgent(),
	}

	resp, err := h.authService.Login(r.Context(), loginReq)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTenantRequired):
			// User needs to select a tenant
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(TenantRequiredResponse{
				Error: ErrorDetail{
					Code:    "tenant_required",
					Message: "User belongs to multiple tenants. Please specify tenant_id.",
					Tenants: ToTenantOptions(resp.Tenants),
				},
			})
			return
		case errors.Is(err, domain.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
			return
		case errors.Is(err, domain.ErrAccountDisabled):
			writeError(w, http.StatusUnauthorized, "account_disabled", "Account is disabled")
			return
		case errors.Is(err, domain.ErrRateLimitExceeded):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{
					Code:       "rate_limit_exceeded",
					Message:    "Too many login attempts. Please try again later.",
					RetryAfter: 60,
				},
			})
			return
		case errors.Is(err, domain.ErrTenantInactive):
			writeError(w, http.StatusUnauthorized, "tenant_inactive", "Tenant is inactive")
			return
		case errors.Is(err, domain.ErrUserNotInTenant):
			writeError(w, http.StatusBadRequest, "invalid_tenant", "User does not belong to this tenant")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			return
		}
	}

	writeJSON(w, http.StatusOK, ToLoginResponse(resp))
}

// Refresh handles POST /refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Refresh token is required")
		return
	}

	refreshReq := service.RefreshRequest{
		RefreshToken: req.RefreshToken,
		IPAddress:    GetClientIP(r),
		UserAgent:    r.UserAgent(),
	}

	tokenPair, err := h.authService.Refresh(r.Context(), refreshReq)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRefreshTokenInvalid):
			writeError(w, http.StatusUnauthorized, "token_invalid", "Refresh token is invalid")
			return
		case errors.Is(err, domain.ErrSessionRevoked):
			writeError(w, http.StatusUnauthorized, "session_revoked", "Session has been revoked")
			return
		case errors.Is(err, domain.ErrTokenExpired):
			writeError(w, http.StatusUnauthorized, "token_expired", "Refresh token has expired")
			return
		case errors.Is(err, domain.ErrAccountDisabled):
			writeError(w, http.StatusUnauthorized, "account_disabled", "Account is disabled")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			return
		}
	}

	writeJSON(w, http.StatusOK, ToTokenResponse(tokenPair))
}

// Logout handles POST /logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Accept empty body - use token from header if available
		req.RefreshToken = ""
	}

	// If no refresh token in body, try to get session from token and revoke
	if req.RefreshToken == "" {
		// Just return success - nothing to revoke
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := h.authService.Logout(r.Context(), req.RefreshToken, GetClientIP(r))
	if err != nil {
		// Log error but return success to client
		// Logout should be idempotent
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /me.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetClaims(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userID, _ := claims.GetUserID()

	// For a full implementation, we'd fetch the user and their tenants
	// For now, return what we have in the token
	resp := MeResponse{
		ID:                userID,
		Email:             claims.Email,
		FirstName:         "", // Would need to fetch from DB
		LastName:          "", // Would need to fetch from DB
		Role:              string(claims.Role),
		TenantID:          claims.TenantID,
		TenantName:        "", // Would need to fetch from DB
		MustResetPassword: false,
		Tenants:           []TenantRoleInfo{},
	}

	writeJSON(w, http.StatusOK, resp)
}

// ChangePassword handles POST /change-password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request, userService *service.UserService) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Current password and new password are required")
		return
	}

	changeReq := service.ChangePasswordRequest{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
		IPAddress:       GetClientIP(r),
	}

	err := userService.ChangePassword(r.Context(), changeReq)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordIncorrect):
			writeError(w, http.StatusBadRequest, "current_password_incorrect", "Current password is incorrect")
			return
		case errors.Is(err, domain.ErrPasswordWeak):
			writeError(w, http.StatusBadRequest, "password_weak", "Password does not meet requirements")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			return
		}
	}

	writeJSON(w, http.StatusOK, MessageResponse{
		Message: "Password changed successfully. All other sessions have been invalidated.",
	})
}

// RequestPasswordReset handles POST /password-reset/request.
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request, userService *service.UserService) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Email is required")
		return
	}

	err := userService.RequestPasswordReset(r.Context(), req.Email, GetClientIP(r))
	if err != nil {
		if errors.Is(err, domain.ErrRateLimitExceeded) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: ErrorDetail{
					Code:       "rate_limit_exceeded",
					Message:    "Please wait before requesting another reset.",
					RetryAfter: 300,
				},
			})
			return
		}
		// Don't reveal other errors - could expose whether email exists
	}

	// Always return success to prevent email enumeration
	w.WriteHeader(http.StatusAccepted)
	writeJSON(w, http.StatusAccepted, MessageResponse{
		Message: "If the email exists, a reset link has been sent.",
	})
}

// CompletePasswordReset handles POST /password-reset/complete.
func (h *AuthHandler) CompletePasswordReset(w http.ResponseWriter, r *http.Request, userService *service.UserService) {
	var req PasswordResetCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Token and new password are required")
		return
	}

	err := userService.CompletePasswordReset(r.Context(), req.Token, req.NewPassword, GetClientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordResetInvalid):
			writeError(w, http.StatusBadRequest, "token_invalid", "Password reset token is invalid")
			return
		case errors.Is(err, domain.ErrPasswordResetExpired):
			writeError(w, http.StatusBadRequest, "token_expired", "Password reset token has expired")
			return
		case errors.Is(err, domain.ErrPasswordResetUsed):
			writeError(w, http.StatusBadRequest, "token_used", "Password reset token has already been used")
			return
		case errors.Is(err, domain.ErrPasswordWeak):
			writeError(w, http.StatusBadRequest, "password_weak", "Password does not meet requirements")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			return
		}
	}

	writeJSON(w, http.StatusOK, MessageResponse{
		Message: "Password has been reset successfully.",
	})
}

// --- Helper Functions ---

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
