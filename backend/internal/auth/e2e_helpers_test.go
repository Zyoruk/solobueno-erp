package auth

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
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository/mock"
	"github.com/solobueno/erp/internal/auth/service"
	"github.com/solobueno/erp/pkg/jwt"
)

// e2eEnv bundles a real HTTP server (real chi router, real services) backed
// by in-memory mock repositories, for full-lifecycle e2e tests that exercise
// every auth/user endpoint over real HTTP rather than calling services directly.
type e2eEnv struct {
	t           *testing.T
	server      *httptest.Server
	client      *http.Client
	userRepo    *mock.MockUserRepository
	tenantRepo  *mock.MockTenantRepository
	roleRepo    *mock.MockUserTenantRoleRepository
	sessionRepo *mock.MockSessionRepository
	resetRepo   *mock.MockPasswordResetRepository
	emailer     *capturingEmailer
}

func setupE2E(t *testing.T) *e2eEnv {
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
	tenantRepo := mock.NewMockTenantRepository()
	roleRepo := mock.NewMockUserTenantRoleRepository()
	resetRepo := mock.NewMockPasswordResetRepository()
	emailer := newCapturingEmailer()

	// Cross-reference the two mock stores the way a real Postgres FK join
	// would: after UserService.Create/UpdateRole writes to roleRepo, reads
	// through userRepo's *WithTenants methods (and ListByTenant) see it.
	userRepo.RolesLookup = func(userID uuid.UUID) []domain.UserTenantRole {
		ptrs, _ := roleRepo.ListByUser(context.Background(), userID)
		roles := make([]domain.UserTenantRole, len(ptrs))
		for i, p := range ptrs {
			roles[i] = *p
		}
		return roles
	}

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
		PasswordReset: resetRepo,
		Emailer:       emailer,
	})

	mux := chi.NewRouter()
	mux.Mount("/", Router(authSvc, userSvc))
	mux.Mount("/users", UserRouter(authSvc, userSvc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return &e2eEnv{
		t: t, server: srv, client: srv.Client(),
		userRepo: userRepo, tenantRepo: tenantRepo, roleRepo: roleRepo,
		sessionRepo: sessionRepo, resetRepo: resetRepo, emailer: emailer,
	}
}

// seedUser creates an active user with a hashed password and a role in tenantID.
func (e *e2eEnv) seedUser(email, password string, tenantID uuid.UUID, role domain.Role) *domain.User {
	e.t.Helper()
	hash, err := service.NewPasswordService().Hash(password)
	if err != nil {
		e.t.Fatalf("failed to hash password: %v", err)
	}
	userID := uuid.New()
	roleID := uuid.New()
	user := &domain.User{
		ID: userID, Email: email, PasswordHash: hash, FirstName: "Test", LastName: "User", IsActive: true,
	}
	e.userRepo.AddUser(user)
	e.roleRepo.AddRole(&domain.UserTenantRole{ID: roleID, UserID: userID, TenantID: tenantID, Role: role})
	return user
}

func (e *e2eEnv) seedTenant(name, slug string) *domain.Tenant {
	e.t.Helper()
	tenant := &domain.Tenant{ID: uuid.New(), Name: name, Slug: slug, IsActive: true}
	e.tenantRepo.AddTenant(tenant)
	return tenant
}

// login performs a real HTTP login and returns the access/refresh tokens
// alongside the raw response (status code, etc.) for the caller to inspect.
func (e *e2eEnv) login(email, password string) (accessToken, refreshToken string, resp *http.Response) {
	e.t.Helper()
	resp = e.do(http.MethodPost, "/login", "", map[string]string{"email": email, "password": password})
	defer resp.Body.Close()
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	return out.AccessToken, out.RefreshToken, resp
}

// do performs a JSON request against the e2e server. token == "" sends no
// Authorization header. body == nil sends no request body.
func (e *e2eEnv) do(method, path, token string, body interface{}) *http.Response {
	e.t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			e.t.Fatalf("failed to marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, e.server.URL+path, reader)
	if err != nil {
		e.t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		e.t.Fatalf("%s %s failed: %v", method, path, err)
	}
	return resp
}

// decodeBody JSON-decodes and closes the response body.
func decodeBody(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

// capturingEmailer is a test double for service.Emailer that records what
// would have been sent, so e2e tests can retrieve tokens/passwords that
// only ever leave the system via email (never in an API response).
type capturingEmailer struct {
	mu            sync.Mutex
	tempPasswords map[string]string
	resetTokens   map[string]string
	tenantLinks   []string
}

func newCapturingEmailer() *capturingEmailer {
	return &capturingEmailer{tempPasswords: map[string]string{}, resetTokens: map[string]string{}}
}

func (e *capturingEmailer) SendTemporaryPassword(ctx context.Context, toEmail, tempPassword string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tempPasswords[toEmail] = tempPassword
	return nil
}

func (e *capturingEmailer) SendTenantLinked(ctx context.Context, toEmail string, tenantID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tenantLinks = append(e.tenantLinks, toEmail)
	return nil
}

func (e *capturingEmailer) SendPasswordReset(ctx context.Context, toEmail, resetToken string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.resetTokens[toEmail] = resetToken
	return nil
}

func (e *capturingEmailer) tempPasswordFor(t *testing.T, email string) string {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.tempPasswords[email]
	if !ok {
		t.Fatalf("no temporary password captured for %s", email)
	}
	return p
}

func (e *capturingEmailer) resetTokenFor(t *testing.T, email string) string {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	tok, ok := e.resetTokens[email]
	if !ok {
		t.Fatalf("no reset token captured for %s", email)
	}
	return tok
}

func (e *capturingEmailer) hasTenantLinkNotification(email string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, addr := range e.tenantLinks {
		if addr == email {
			return true
		}
	}
	return false
}

var _ service.Emailer = (*capturingEmailer)(nil)
