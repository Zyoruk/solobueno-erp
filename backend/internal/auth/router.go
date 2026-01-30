package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/handler"
	"github.com/solobueno/erp/internal/auth/service"
)

// Router creates and configures the auth router.
func Router(authService *service.AuthService, userService *service.UserService) chi.Router {
	r := chi.NewRouter()

	authHandler := handler.NewAuthHandler(authService)
	middleware := handler.NewAuthMiddleware(authService)

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/password-reset/request", func(w http.ResponseWriter, req *http.Request) {
			authHandler.RequestPasswordReset(w, req, userService)
		})
		r.Post("/password-reset/complete", func(w http.ResponseWriter, req *http.Request) {
			authHandler.CompletePasswordReset(w, req, userService)
		})
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		// Auth endpoints
		r.Post("/logout", authHandler.Logout)
		r.Get("/me", authHandler.Me)
		r.Post("/change-password", func(w http.ResponseWriter, req *http.Request) {
			authHandler.ChangePassword(w, req, userService)
		})
	})

	return r
}

// UserRouter creates and configures the user management router.
func UserRouter(authService *service.AuthService, userService *service.UserService) chi.Router {
	r := chi.NewRouter()

	userHandler := handler.NewUserHandler(userService)
	middleware := handler.NewAuthMiddleware(authService)

	// All user routes require authentication
	r.Use(middleware.RequireAuth)

	// Routes requiring Manager+ role
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.RoleManager))

		r.Post("/", userHandler.Create)
		r.Get("/", userHandler.List)
		r.Get("/{id}", userHandler.Get)
		r.Patch("/{id}", userHandler.Update)
		r.Patch("/{id}/role", userHandler.UpdateRole)
	})

	return r
}
