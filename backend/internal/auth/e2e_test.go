package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository/mock"
	"github.com/solobueno/erp/internal/auth/service"
	"github.com/solobueno/erp/pkg/jwt"
)

// TestE2E_LoginUseRefreshLogout is T082: the full login -> use -> refresh ->
// logout flow, driven over real HTTP against the real routers (mock repos
// stand in for Postgres — main.go does not yet wire a live DB/router, see
// the tasks.md note this test's addition documents).
func TestE2E_LoginUseRefreshLogout(t *testing.T) {
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
	tenantRepo := mock.NewMockTenantRepository()
	roleRepo := mock.NewMockUserTenantRoleRepository()

	authSvc := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		EventRepo:    mock.NewMockAuthEventRepository(),
		TenantRepo:   tenantRepo,
		RoleRepo:     roleRepo,
		TokenService: tokenSvc,
	})
	userSvc := service.NewUserService(service.UserServiceConfig{
		UserRepo:      userRepo,
		RoleRepo:      roleRepo,
		SessionRepo:   sessionRepo,
		EventRepo:     mock.NewMockAuthEventRepository(),
		PasswordReset: mock.NewMockPasswordResetRepository(),
	})

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, err := service.NewPasswordService().Hash("Password123!")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	userRepo.AddUser(&domain.User{
		ID: userID, Email: "e2e@example.com", PasswordHash: passwordHash, IsActive: true,
		TenantRoles: []domain.UserTenantRole{{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleManager}},
	})
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, Name: "Acme", Slug: "acme", IsActive: true})

	mux := chi.NewRouter()
	mux.Mount("/", Router(authSvc, userSvc))
	mux.Mount("/users", UserRouter(authSvc, userSvc))

	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := srv.Client()

	// 1. Login
	loginBody, _ := json.Marshal(map[string]string{"email": "e2e@example.com", "password": "Password123!"})
	resp, err := client.Post(srv.URL+"/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var loginResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	resp.Body.Close()
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Fatal("login did not return tokens")
	}

	// 2. Use: GET /me with the access token
	meReq, _ := http.NewRequest("GET", srv.URL+"/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	resp, err = client.Do(meReq)
	if err != nil {
		t.Fatalf("/me request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/me status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// 3. Refresh
	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": loginResp.RefreshToken})
	resp, err = client.Post(srv.URL+"/refresh", "application/json", bytes.NewReader(refreshBody))
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var refreshResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(resp.Body).Decode(&refreshResp)
	resp.Body.Close()
	if refreshResp.AccessToken == "" || refreshResp.RefreshToken == loginResp.RefreshToken {
		t.Fatal("refresh did not rotate tokens")
	}

	// The old refresh token must now be rejected (rotation).
	resp, err = client.Post(srv.URL+"/refresh", "application/json", bytes.NewReader(refreshBody))
	if err != nil {
		t.Fatalf("stale refresh request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("stale refresh token status = %d, want %d (token rotation not enforced)", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()

	// 4. Logout with the rotated refresh token (protected route: needs the
	// current access token too).
	logoutBody, _ := json.Marshal(map[string]string{"refresh_token": refreshResp.RefreshToken})
	logoutReq, _ := http.NewRequest("POST", srv.URL+"/logout", bytes.NewReader(logoutBody))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutReq.Header.Set("Authorization", "Bearer "+refreshResp.AccessToken)
	resp, err = client.Do(logoutReq)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	resp.Body.Close()

	// The logged-out refresh token must be rejected.
	resp, err = client.Post(srv.URL+"/refresh", "application/json", bytes.NewReader(logoutBody))
	if err != nil {
		t.Fatalf("post-logout refresh request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("post-logout refresh status = %d, want %d (session not revoked)", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()
}
