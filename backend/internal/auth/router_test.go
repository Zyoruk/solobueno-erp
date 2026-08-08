package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/solobueno/erp/internal/auth/repository/mock"
	"github.com/solobueno/erp/internal/auth/service"
	"github.com/solobueno/erp/pkg/jwt"
)

// publicRoutes lists the only endpoints allowed to skip authentication:
// login/refresh (that's how you get a token) and password reset (used by
// someone who, by definition, can't log in yet). SC-003 requires 100% of
// every other endpoint to enforce auth.
var publicRoutes = map[string]bool{
	"POST /login":                   true,
	"POST /refresh":                 true,
	"POST /password-reset/request":  true,
	"POST /password-reset/complete": true,
}

func testServices(t *testing.T) (*service.AuthService, *service.UserService) {
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
	userSvc := service.NewUserService(service.UserServiceConfig{
		UserRepo:      mock.NewMockUserRepository(),
		RoleRepo:      mock.NewMockUserTenantRoleRepository(),
		SessionRepo:   mock.NewMockSessionRepository(),
		EventRepo:     mock.NewMockAuthEventRepository(),
		PasswordReset: mock.NewMockPasswordResetRepository(),
	})

	return authSvc, userSvc
}

// TestRouteAuthCoverage walks every registered route in both the auth and
// user routers and, for anything not explicitly public, fires a request
// with no Authorization header. Each must come back 401 — proving the
// route actually goes through RequireAuth rather than just trusting that a
// r.Use() call was added correctly (SC-003, SC-004).
func TestRouteAuthCoverage(t *testing.T) {
	authSvc, userSvc := testServices(t)

	routers := map[string]chi.Router{
		"auth": Router(authSvc, userSvc),
		"user": UserRouter(authSvc, userSvc),
	}

	checked := 0
	for name, r := range routers {
		err := chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			key := method + " " + route
			if publicRoutes[key] {
				return nil
			}

			wrapped := handler
			for i := len(middlewares) - 1; i >= 0; i-- {
				wrapped = middlewares[i](wrapped)
			}

			req := httptest.NewRequest(method, route, nil)
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s router %s: expected 401 without auth, got %d (route not protected by RequireAuth)", name, key, w.Code)
			}
			checked++
			return nil
		})
		if err != nil {
			t.Fatalf("chi.Walk failed for %s router: %v", name, err)
		}
	}

	if checked == 0 {
		t.Fatal("no protected routes were checked — route list may have changed, update publicRoutes/test")
	}
}
