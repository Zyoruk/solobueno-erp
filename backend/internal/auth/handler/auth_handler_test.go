package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
)

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	handler := NewAuthHandler(nil)

	tests := []struct {
		name    string
		body    LoginRequest
		wantErr string
	}{
		{
			name:    "missing email",
			body:    LoginRequest{Password: "password"},
			wantErr: "invalid_request",
		},
		{
			name:    "missing password",
			body:    LoginRequest{Email: "test@example.com"},
			wantErr: "invalid_request",
		},
		{
			name:    "missing both",
			body:    LoginRequest{},
			wantErr: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
			}

			var errResp ErrorResponse
			json.NewDecoder(w.Body).Decode(&errResp)
			if errResp.Error.Code != tt.wantErr {
				t.Errorf("Error code = %q, want %q", errResp.Error.Code, tt.wantErr)
			}
		})
	}
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Refresh_MissingToken(t *testing.T) {
	handler := NewAuthHandler(nil)

	bodyBytes, _ := json.Marshal(RefreshRequest{RefreshToken: ""})
	req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Refresh_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	req := httptest.NewRequest("POST", "/refresh", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Logout_EmptyBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	req := httptest.NewRequest("POST", "/logout", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestAuthHandler_Logout_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	req := httptest.NewRequest("POST", "/logout", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	// Should return success even with invalid body
	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	handler := NewAuthHandler(nil)

	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	handler.Me(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Me_WithClaims(t *testing.T) {
	handler := NewAuthHandler(nil)

	userID := uuid.New()
	tenantID := uuid.New()
	claims := &domain.Claims{
		TenantID: tenantID,
		Role:     domain.RoleManager,
		Email:    "test@example.com",
	}
	claims.Subject = userID.String()

	ctx := context.WithValue(context.Background(), UserContextKey, claims)
	req := httptest.NewRequest("GET", "/me", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Me(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp MeResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", resp.Email, "test@example.com")
	}
	if resp.Role != "manager" {
		t.Errorf("Role = %q, want %q", resp.Role, "manager")
	}
}

func TestAuthHandler_ChangePassword_Unauthorized(t *testing.T) {
	handler := NewAuthHandler(nil)

	bodyBytes, _ := json.Marshal(ChangePasswordRequest{
		CurrentPassword: "old",
		NewPassword:     "new",
	})
	req := httptest.NewRequest("POST", "/change-password", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.ChangePassword(w, req, nil)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_ChangePassword_MissingFields(t *testing.T) {
	handler := NewAuthHandler(nil)

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), UserIDContextKey, userID)

	tests := []struct {
		name string
		body ChangePasswordRequest
	}{
		{
			name: "missing current password",
			body: ChangePasswordRequest{NewPassword: "newpassword"},
		},
		{
			name: "missing new password",
			body: ChangePasswordRequest{CurrentPassword: "oldpassword"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/change-password", bytes.NewReader(bodyBytes)).WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ChangePassword(w, req, nil)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestAuthHandler_RequestPasswordReset_MissingEmail(t *testing.T) {
	handler := NewAuthHandler(nil)

	bodyBytes, _ := json.Marshal(PasswordResetRequest{Email: ""})
	req := httptest.NewRequest("POST", "/password-reset/request", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.RequestPasswordReset(w, req, nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_RequestPasswordReset_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	req := httptest.NewRequest("POST", "/password-reset/request", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.RequestPasswordReset(w, req, nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_CompletePasswordReset_MissingFields(t *testing.T) {
	handler := NewAuthHandler(nil)

	tests := []struct {
		name string
		body PasswordResetCompleteRequest
	}{
		{
			name: "missing token",
			body: PasswordResetCompleteRequest{NewPassword: "newpassword"},
		},
		{
			name: "missing password",
			body: PasswordResetCompleteRequest{Token: "reset-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/password-reset/complete", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler.CompletePasswordReset(w, req, nil)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestAuthHandler_CompletePasswordReset_InvalidBody(t *testing.T) {
	handler := NewAuthHandler(nil)

	req := httptest.NewRequest("POST", "/password-reset/complete", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.CompletePasswordReset(w, req, nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
