package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository/mock"
	"github.com/solobueno/erp/internal/auth/service"
	"github.com/solobueno/erp/pkg/jwt"
)

// setupAuthMiddleware builds a real AuthService (mock repos, in-memory RSA
// keypair) so RequireAuth exercises actual token validation instead of nils.
func setupAuthMiddleware(t *testing.T) (*AuthMiddleware, *service.TokenService) {
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

	authSvc := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     mock.NewMockUserRepository(),
		SessionRepo:  mock.NewMockSessionRepository(),
		EventRepo:    mock.NewMockAuthEventRepository(),
		TenantRepo:   mock.NewMockTenantRepository(),
		RoleRepo:     mock.NewMockUserTenantRoleRepository(),
		TokenService: tokenSvc,
	})

	return NewAuthMiddleware(authSvc), tokenSvc
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	mw, tokenSvc := setupAuthMiddleware(t)

	user := &domain.User{ID: uuid.New(), Email: "user@example.com"}
	tenantID := uuid.New()
	pair, _, err := tokenSvc.GenerateTokenPair(user, tenantID, domain.RoleManager)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims, ok := GetClaims(r.Context())
		if !ok || claims.Email != "user@example.com" {
			t.Error("expected claims in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()

	mw.RequireAuth(next).ServeHTTP(w, req)

	if !called {
		t.Error("next handler should be called for a valid token")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_RequireAuth_MissingToken(t *testing.T) {
	mw, _ := setupAuthMiddleware(t)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	mw.RequireAuth(next).ServeHTTP(w, req)

	if called {
		t.Error("next handler should not be called without a token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	mw, _ := setupAuthMiddleware(t)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()

	mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	mw, _ := setupAuthMiddleware(t)

	tests := []struct {
		name     string
		role     domain.Role
		minRole  domain.Role
		wantCode int
	}{
		{"sufficient role", domain.RoleOwner, domain.RoleManager, http.StatusOK},
		{"exact role", domain.RoleManager, domain.RoleManager, http.StatusOK},
		{"insufficient role", domain.RoleWaiter, domain.RoleManager, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			handler := mw.RequireRole(tt.minRole)(next)

			ctx := context.WithValue(context.Background(), RoleContextKey, tt.role)
			req := httptest.NewRequest("GET", "/admin", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestAuthMiddleware_RequireRole_Unauthenticated(t *testing.T) {
	mw, _ := setupAuthMiddleware(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := mw.RequireRole(domain.RoleManager)(next)

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RequireAnyRole(t *testing.T) {
	mw, _ := setupAuthMiddleware(t)

	tests := []struct {
		name     string
		role     domain.Role
		wantCode int
	}{
		{"allowed role", domain.RoleCashier, http.StatusOK},
		{"another allowed role", domain.RoleWaiter, http.StatusOK},
		{"disallowed role", domain.RoleKitchen, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			handler := mw.RequireAnyRole(domain.RoleCashier, domain.RoleWaiter)(next)

			ctx := context.WithValue(context.Background(), RoleContextKey, tt.role)
			req := httptest.NewRequest("GET", "/orders", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestAuthMiddleware_RequireAnyRole_Unauthenticated(t *testing.T) {
	mw, _ := setupAuthMiddleware(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := mw.RequireAnyRole(domain.RoleCashier)(next)

	req := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantErr   bool
	}{
		{
			name:      "valid bearer token",
			header:    "Bearer abc123",
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name:      "lowercase bearer",
			header:    "bearer abc123",
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name:      "missing header",
			header:    "",
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "wrong scheme",
			header:    "Basic abc123",
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "no token after bearer",
			header:    "Bearer",
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token, err := extractBearerToken(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token != tt.wantToken {
				t.Errorf("extractBearerToken() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		trustProxy bool
		xff        string
		xRealIP    string
		remoteAddr string
		want       string
	}{
		{
			name:       "X-Forwarded-For single, proxy trusted",
			trustProxy: true,
			xff:        "192.168.1.1",
			remoteAddr: "10.0.0.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For multiple, proxy trusted",
			trustProxy: true,
			xff:        "192.168.1.1, 10.0.0.2, 172.16.0.1",
			remoteAddr: "10.0.0.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "X-Real-IP, proxy trusted",
			trustProxy: true,
			xRealIP:    "192.168.1.2",
			remoteAddr: "10.0.0.1:12345",
			want:       "192.168.1.2",
		},
		{
			name:       "X-Forwarded-For ignored when proxy not trusted",
			trustProxy: false,
			xff:        "192.168.1.1",
			remoteAddr: "10.0.0.1:12345",
			want:       "10.0.0.1",
		},
		{
			name:       "RemoteAddr only",
			remoteAddr: "10.0.0.1:12345",
			want:       "10.0.0.1",
		},
		{
			name:       "RemoteAddr no port",
			remoteAddr: "10.0.0.1",
			want:       "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.trustProxy {
				t.Setenv("TRUST_PROXY_HEADERS", "true")
			}
			req := httptest.NewRequest("GET", "/", nil)
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			req.RemoteAddr = tt.remoteAddr

			if got := GetClientIP(req); got != tt.want {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextHelpers(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	role := domain.RoleManager

	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDContextKey, userID)
	ctx = context.WithValue(ctx, TenantIDContextKey, tenantID)
	ctx = context.WithValue(ctx, RoleContextKey, role)

	// Test GetUserID
	gotUserID, ok := GetUserID(ctx)
	if !ok {
		t.Error("GetUserID() should return true")
	}
	if gotUserID != userID {
		t.Errorf("GetUserID() = %v, want %v", gotUserID, userID)
	}

	// Test GetTenantID
	gotTenantID, ok := GetTenantID(ctx)
	if !ok {
		t.Error("GetTenantID() should return true")
	}
	if gotTenantID != tenantID {
		t.Errorf("GetTenantID() = %v, want %v", gotTenantID, tenantID)
	}

	// Test GetRole
	gotRole, ok := GetRole(ctx)
	if !ok {
		t.Error("GetRole() should return true")
	}
	if gotRole != role {
		t.Errorf("GetRole() = %v, want %v", gotRole, role)
	}

	// Test with empty context
	emptyCtx := context.Background()
	_, ok = GetUserID(emptyCtx)
	if ok {
		t.Error("GetUserID() on empty context should return false")
	}
}

func TestGetClaims(t *testing.T) {
	claims := &domain.Claims{
		TenantID: uuid.New(),
		Role:     domain.RoleAdmin,
		Email:    "test@example.com",
	}

	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	gotClaims, ok := GetClaims(ctx)
	if !ok {
		t.Error("GetClaims() should return true")
	}
	if gotClaims.Email != claims.Email {
		t.Errorf("GetClaims().Email = %q, want %q", gotClaims.Email, claims.Email)
	}

	// Test with empty context
	_, ok = GetClaims(context.Background())
	if ok {
		t.Error("GetClaims() on empty context should return false")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"message": "hello"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	expected := `{"message":"hello"}`
	// Note: json.Encoder adds a newline
	if w.Body.String() != expected+"\n" {
		t.Errorf("Body = %q, want %q", w.Body.String(), expected)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "invalid_request", "Something went wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	// Check that body contains error code
	body := w.Body.String()
	if body == "" {
		t.Error("Body should not be empty")
	}
}
