package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/solobueno/erp/internal/auth/repository"
	"github.com/solobueno/erp/internal/auth/service"
	"github.com/solobueno/erp/pkg/jwt"
	"gorm.io/gorm"
)

// Module represents the auth module with all its components.
type Module struct {
	AuthService *service.AuthService
	UserService *service.UserService
	AuthRouter  chi.Router
	UserRouter  chi.Router
}

// ModuleConfig holds configuration for the auth module.
type ModuleConfig struct {
	DB         *gorm.DB
	KeyManager *jwt.KeyManager
	JWTConfig  jwt.TokenGeneratorConfig
}

// NewModule creates and initializes the auth module.
func NewModule(cfg ModuleConfig) (*Module, error) {
	// Create repositories
	userRepo := repository.NewGormUserRepository(cfg.DB)
	sessionRepo := repository.NewGormSessionRepository(cfg.DB)
	eventRepo := repository.NewGormAuthEventRepository(cfg.DB)
	tenantRepo := repository.NewGormTenantRepository(cfg.DB)
	roleRepo := repository.NewGormUserTenantRoleRepository(cfg.DB)
	passwordResetRepo := repository.NewGormPasswordResetRepository(cfg.DB)

	// Create token service
	tokenService := service.NewTokenService(cfg.KeyManager, cfg.JWTConfig)

	// Create rate limiters
	loginRateLimiter := service.NewMemoryRateLimiter(service.DefaultLoginRateLimiterConfig())
	resetRateLimiter := service.NewMemoryRateLimiter(service.DefaultPasswordResetRateLimiterConfig())

	// Create services
	authService := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     userRepo,
		SessionRepo:  sessionRepo,
		EventRepo:    eventRepo,
		TenantRepo:   tenantRepo,
		RoleRepo:     roleRepo,
		TokenService: tokenService,
		RateLimiter:  loginRateLimiter,
	})

	userService := service.NewUserService(service.UserServiceConfig{
		UserRepo:         userRepo,
		RoleRepo:         roleRepo,
		SessionRepo:      sessionRepo,
		EventRepo:        eventRepo,
		PasswordReset:    passwordResetRepo,
		ResetRateLimiter: resetRateLimiter,
	})

	// Create routers
	authRouter := Router(authService, userService)
	userRouter := UserRouter(authService, userService)

	return &Module{
		AuthService: authService,
		UserService: userService,
		AuthRouter:  authRouter,
		UserRouter:  userRouter,
	}, nil
}

// RegisterRoutes registers the auth module routes with a parent router.
func (m *Module) RegisterRoutes(r chi.Router) {
	r.Mount("/api/v1/auth", m.AuthRouter)
	r.Mount("/api/v1/users", m.UserRouter)
}
