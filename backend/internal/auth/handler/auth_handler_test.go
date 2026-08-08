package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository/mock"
	"github.com/solobueno/erp/internal/auth/service"
	"github.com/solobueno/erp/pkg/jwt"
)

// setupWiredAuthHandler builds an AuthHandler and UserHandler backed by a
// real AuthService/UserService (mock repos, in-memory RSA keypair) so
// success and service-error branches are exercised, not just validation.
func setupWiredAuthHandler(t *testing.T) (*AuthHandler, *service.UserService, *service.TokenService, *mock.MockUserRepository, *mock.MockTenantRepository, *mock.MockUserTenantRoleRepository, *mock.MockSessionRepository) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyBytes})

	km := jwt.NewKeyManager()
	if err := km.LoadPrivateKeyFromPEM(privatePEM); err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}
	if err := km.LoadPublicKeyFromPEM(publicPEM); err != nil {
		t.Fatalf("failed to load public key: %v", err)
	}
	tokenSvc := service.NewTokenService(km, jwt.DefaultTokenGeneratorConfig())

	userRepo := mock.NewMockUserRepository()
	sessionRepo := mock.NewMockSessionRepository()
	eventRepo := mock.NewMockAuthEventRepository()
	tenantRepo := mock.NewMockTenantRepository()
	roleRepo := mock.NewMockUserTenantRoleRepository()
	passwordResetRepo := mock.NewMockPasswordResetRepository()

	authSvc := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		EventRepo:    eventRepo,
		TenantRepo:   tenantRepo,
		RoleRepo:     roleRepo,
		TokenService: tokenSvc,
	})

	userSvc := service.NewUserService(service.UserServiceConfig{
		UserRepo:      userRepo,
		RoleRepo:      roleRepo,
		SessionRepo:   sessionRepo,
		EventRepo:     eventRepo,
		PasswordReset: passwordResetRepo,
	})

	return NewAuthHandler(authSvc), userSvc, tokenSvc, userRepo, tenantRepo, roleRepo, sessionRepo
}

func TestAuthHandler_Login_Success(t *testing.T) {
	h, _, _, userRepo, tenantRepo, _, _ := setupWiredAuthHandler(t)

	passwordHash, _ := service.NewPasswordService().Hash("Password123!")
	tenantID := uuid.New()
	userID := uuid.New()
	user := &domain.User{
		ID: userID, Email: "login@example.com", PasswordHash: passwordHash, IsActive: true,
		TenantRoles: []domain.UserTenantRole{{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleManager}},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, Name: "Acme", Slug: "acme", IsActive: true})

	body, _ := json.Marshal(LoginRequest{Email: "login@example.com", Password: "Password123!"})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp LoginResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.AccessToken == "" || resp.User.Email != "login@example.com" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	h, _, _, userRepo, _, _, _ := setupWiredAuthHandler(t)

	passwordHash, _ := service.NewPasswordService().Hash("Password123!")
	userRepo.AddUser(&domain.User{ID: uuid.New(), Email: "wrong@example.com", PasswordHash: passwordHash, IsActive: true})

	body, _ := json.Marshal(LoginRequest{Email: "wrong@example.com", Password: "WrongPassword!"})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestAuthHandler_Login_AccountDisabled(t *testing.T) {
	h, _, _, userRepo, _, _, _ := setupWiredAuthHandler(t)

	passwordHash, _ := service.NewPasswordService().Hash("Password123!")
	userRepo.AddUser(&domain.User{ID: uuid.New(), Email: "disabled@example.com", PasswordHash: passwordHash, IsActive: false})

	body, _ := json.Marshal(LoginRequest{Email: "disabled@example.com", Password: "Password123!"})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestAuthHandler_Login_AccountLocked(t *testing.T) {
	h, _, _, userRepo, tenantRepo, _, _ := setupWiredAuthHandler(t)

	passwordHash, _ := service.NewPasswordService().Hash("CorrectPassword123!")
	tenantID := uuid.New()
	userID := uuid.New()
	user := &domain.User{
		ID: userID, Email: "locked@example.com", PasswordHash: passwordHash, IsActive: true,
		TenantRoles: []domain.UserTenantRole{{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleManager}},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, Name: "Acme", Slug: "acme", IsActive: true})

	for i := 0; i < 5; i++ {
		body, _ := json.Marshal(LoginRequest{Email: "locked@example.com", Password: "WrongPassword!"})
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.Login(w, req)
	}

	body, _ := json.Marshal(LoginRequest{Email: "locked@example.com", Password: "CorrectPassword123!"})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusLocked {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusLocked, w.Body.String())
	}
	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "account_locked" || resp.Error.LockedUntil == nil {
		t.Errorf("unexpected error response: %+v", resp.Error)
	}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	h, _, tokenSvc, userRepo, tenantRepo, _, sessionRepo := setupWiredAuthHandler(t)

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	user := &domain.User{
		ID: userID, Email: "refresh@example.com", IsActive: true,
		TenantRoles: []domain.UserTenantRole{{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleManager}},
	}
	userRepo.AddUser(user)
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, Name: "Acme", Slug: "acme", IsActive: true})
	sessionRepo.Create(ctx, &domain.Session{ID: uuid.New(), UserID: user.ID, TenantID: tenantID, RefreshToken: tokenSvc.HashRefreshToken("plain-refresh-token"), ExpiresAt: time.Now().Add(time.Hour)})

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "plain-refresh-token"})
	req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	h, _, _, _, _, _, _ := setupWiredAuthHandler(t)

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "nonexistent-token"})
	req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestAuthHandler_Logout_RevokesSession(t *testing.T) {
	h, _, tokenSvc, userRepo, _, _, sessionRepo := setupWiredAuthHandler(t)

	ctx := context.Background()
	user := &domain.User{ID: uuid.New(), Email: "logout@example.com", IsActive: true}
	userRepo.AddUser(user)
	sessionID := uuid.New()
	sessionRepo.Create(ctx, &domain.Session{ID: sessionID, UserID: user.ID, TenantID: uuid.New(), RefreshToken: tokenSvc.HashRefreshToken("logout-token"), ExpiresAt: time.Now().Add(time.Hour)})

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "logout-token"})
	req := httptest.NewRequest("POST", "/logout", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
	session, _ := sessionRepo.FindByID(context.Background(), sessionID)
	if session.RevokedAt == nil {
		t.Error("session should be revoked after logout")
	}
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	_, userSvc, _, userRepo, _, _, _ := setupWiredAuthHandler(t)
	h := NewAuthHandler(nil)

	passwordHash, _ := service.NewPasswordService().Hash("OldPassword123!")
	user := &domain.User{ID: uuid.New(), Email: "changepw@example.com", PasswordHash: passwordHash, IsActive: true}
	userRepo.AddUser(user)

	body, _ := json.Marshal(ChangePasswordRequest{CurrentPassword: "OldPassword123!", NewPassword: "NewPassword456!"})
	ctx := context.WithValue(context.Background(), UserIDContextKey, user.ID)
	req := httptest.NewRequest("POST", "/change-password", bytes.NewReader(body)).WithContext(ctx)
	w := httptest.NewRecorder()

	h.ChangePassword(w, req, userSvc)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAuthHandler_ChangePassword_IncorrectCurrent(t *testing.T) {
	_, userSvc, _, userRepo, _, _, _ := setupWiredAuthHandler(t)
	h := NewAuthHandler(nil)

	passwordHash, _ := service.NewPasswordService().Hash("OldPassword123!")
	user := &domain.User{ID: uuid.New(), Email: "wrongpw@example.com", PasswordHash: passwordHash, IsActive: true}
	userRepo.AddUser(user)

	body, _ := json.Marshal(ChangePasswordRequest{CurrentPassword: "WrongOld!", NewPassword: "NewPassword456!"})
	ctx := context.WithValue(context.Background(), UserIDContextKey, user.ID)
	req := httptest.NewRequest("POST", "/change-password", bytes.NewReader(body)).WithContext(ctx)
	w := httptest.NewRecorder()

	h.ChangePassword(w, req, userSvc)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestAuthHandler_RequestPasswordReset_Success(t *testing.T) {
	_, userSvc, _, userRepo, _, _, _ := setupWiredAuthHandler(t)
	h := NewAuthHandler(nil)

	userRepo.AddUser(&domain.User{ID: uuid.New(), Email: "reset@example.com", IsActive: true})

	body, _ := json.Marshal(PasswordResetRequest{Email: "reset@example.com"})
	req := httptest.NewRequest("POST", "/password-reset/request", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RequestPasswordReset(w, req, userSvc)

	if w.Code != http.StatusAccepted {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusAccepted, w.Body.String())
	}
}

func TestAuthHandler_CompletePasswordReset_InvalidToken(t *testing.T) {
	_, userSvc, _, _, _, _, _ := setupWiredAuthHandler(t)
	h := NewAuthHandler(nil)

	body, _ := json.Marshal(PasswordResetCompleteRequest{Token: "nonexistent", NewPassword: "NewPassword456!"})
	req := httptest.NewRequest("POST", "/password-reset/complete", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.CompletePasswordReset(w, req, userSvc)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

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
