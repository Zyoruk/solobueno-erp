package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/service"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create handles POST /users.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	callerRole, ok := GetRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	callerID, _ := GetUserID(r.Context())
	tenantID, _ := GetTenantID(r.Context())

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.FirstName == "" || req.LastName == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Email, first name, and last name are required")
		return
	}

	if !req.Role.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_role", "Invalid role specified")
		return
	}

	createReq := service.CreateUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		TenantID:  tenantID,
		Role:      req.Role,
		CreatedBy: callerID,
		IPAddress: GetClientIP(r),
	}

	resp, err := h.userService.Create(r.Context(), createReq, callerRole)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCannotAssignRole):
			writeError(w, http.StatusForbidden, "insufficient_role", "Cannot create users with role equal or higher than your own")
			return
		case errors.Is(err, domain.ErrEmailExists):
			writeError(w, http.StatusBadRequest, "email_exists", "Email already registered")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			return
		}
	}

	writeJSON(w, http.StatusCreated, CreateUserResponse{
		ID:                resp.User.ID,
		Email:             resp.User.Email,
		FirstName:         resp.User.FirstName,
		LastName:          resp.User.LastName,
		Role:              string(createReq.Role),
		TemporaryPassword: resp.TemporaryPassword,
		MustResetPassword: true,
		CreatedAt:         resp.User.CreatedAt,
	})
}

// List handles GET /users.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.userService.List(r.Context(), tenantID, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *ToUserResponse(user)
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, UserListResponse{
		Data: userResponses,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Get handles GET /users/{id}.
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	writeJSON(w, http.StatusOK, ToUserResponse(user))
}

// Update handles PATCH /users/{id}.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	callerRole, ok := GetRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	callerID, _ := GetUserID(r.Context())
	tenantID, _ := GetTenantID(r.Context())

	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	updateReq := service.UpdateRequest{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  req.IsActive,
		UpdatedBy: callerID,
		TenantID:  tenantID,
		IPAddress: GetClientIP(r),
	}

	user, err := h.userService.Update(r.Context(), updateReq, callerRole)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "not_found", "User not found")
			return
		case errors.Is(err, domain.ErrCannotManageRole):
			writeError(w, http.StatusForbidden, "insufficient_role", "Cannot manage users with this role")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			return
		}
	}

	writeJSON(w, http.StatusOK, ToUserResponse(user))
}

// UpdateRole handles PATCH /users/{id}/role.
func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	callerRole, ok := GetRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	callerID, _ := GetUserID(r.Context())
	tenantID, _ := GetTenantID(r.Context())

	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if !req.Role.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_role", "Invalid role specified")
		return
	}

	updateReq := service.UpdateRoleRequest{
		UserID:    userID,
		TenantID:  tenantID,
		NewRole:   req.Role,
		UpdatedBy: callerID,
		IPAddress: GetClientIP(r),
	}

	err = h.userService.UpdateRole(r.Context(), updateReq, callerRole)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotInTenant):
			writeError(w, http.StatusNotFound, "not_found", "User not found in this tenant")
			return
		case errors.Is(err, domain.ErrCannotAssignRole):
			writeError(w, http.StatusForbidden, "insufficient_role", "Cannot assign role equal or higher than your own")
			return
		case errors.Is(err, domain.ErrCannotManageRole):
			writeError(w, http.StatusForbidden, "insufficient_role", "Cannot manage users with this role")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
			return
		}
	}

	// Fetch updated user to return
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	writeJSON(w, http.StatusOK, ToUserResponse(user))
}
