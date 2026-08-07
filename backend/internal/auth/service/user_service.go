package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository"
)

// UserService handles user management operations.
type UserService struct {
	userRepo        repository.UserRepository
	roleRepo        repository.UserTenantRoleRepository
	sessionRepo     repository.SessionRepository
	eventRepo       repository.AuthEventRepository
	passwordReset   repository.PasswordResetRepository
	passwordSvc     *PasswordService
	resetRateLimiter RateLimiter
}

// UserServiceConfig holds configuration for UserService.
type UserServiceConfig struct {
	UserRepo         repository.UserRepository
	RoleRepo         repository.UserTenantRoleRepository
	SessionRepo      repository.SessionRepository
	EventRepo        repository.AuthEventRepository
	PasswordReset    repository.PasswordResetRepository
	ResetRateLimiter RateLimiter
}

// NewUserService creates a new UserService.
func NewUserService(cfg UserServiceConfig) *UserService {
	return &UserService{
		userRepo:        cfg.UserRepo,
		roleRepo:        cfg.RoleRepo,
		sessionRepo:     cfg.SessionRepo,
		eventRepo:       cfg.EventRepo,
		passwordReset:   cfg.PasswordReset,
		passwordSvc:     NewPasswordService(),
		resetRateLimiter: cfg.ResetRateLimiter,
	}
}

// CreateUserRequest contains the data needed to create a user.
type CreateUserRequest struct {
	Email     string
	FirstName string
	LastName  string
	TenantID  uuid.UUID
	Role      domain.Role
	CreatedBy uuid.UUID // User performing the creation
	IPAddress string
}

// CreateUserResponse contains the result of user creation.
type CreateUserResponse struct {
	User              *domain.User
	TemporaryPassword string
}

// Create creates a new user with a temporary password.
func (s *UserService) Create(ctx context.Context, req CreateUserRequest, callerRole domain.Role) (*CreateUserResponse, error) {
	// Check if caller can assign this role
	if !callerRole.CanAssign(req.Role) {
		return nil, domain.ErrCannotAssignRole
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailExists
	}

	// Generate temporary password
	tempPassword, err := s.passwordSvc.GenerateTemporaryPassword()
	if err != nil {
		return nil, err
	}

	// Hash the password
	passwordHash, err := s.passwordSvc.Hash(tempPassword)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
		MustResetPwd: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create tenant role assignment
	roleAssignment := &domain.UserTenantRole{
		ID:       uuid.New(),
		UserID:   user.ID,
		TenantID: req.TenantID,
		Role:     req.Role,
	}

	if err := s.roleRepo.Create(ctx, roleAssignment); err != nil {
		// Rollback user creation would be ideal here with a transaction
		return nil, err
	}

	// Log event
	s.logEvent(ctx, domain.EventAccountCreated, &user.ID, &req.TenantID, req.IPAddress, "", map[string]interface{}{
		"created_by": req.CreatedBy,
		"role":       req.Role,
	})

	return &CreateUserResponse{
		User:              user,
		TemporaryPassword: tempPassword,
	}, nil
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByIDWithTenants(ctx, id)
}

// UpdateRequest contains the data for updating a user.
type UpdateRequest struct {
	UserID    uuid.UUID
	FirstName *string
	LastName  *string
	IsActive  *bool
	UpdatedBy uuid.UUID
	TenantID  uuid.UUID
	IPAddress string
}

// Update updates a user's profile.
func (s *UserService) Update(ctx context.Context, req UpdateRequest, callerRole domain.Role) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Get target user's role in the tenant
	targetRole := user.GetRoleForTenant(req.TenantID)
	if targetRole != "" && !callerRole.CanManage(targetRole) {
		return nil, domain.ErrCannotManageRole
	}

	// Apply updates
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive

		// Log activation/deactivation
		eventType := domain.EventAccountEnabled
		if !*req.IsActive {
			eventType = domain.EventAccountDisabled
			// Revoke all sessions when disabling
			_ = s.sessionRepo.RevokeAllForUser(ctx, user.ID)
		}
		s.logEvent(ctx, eventType, &user.ID, &req.TenantID, req.IPAddress, "", map[string]interface{}{
			"updated_by": req.UpdatedBy,
		})
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateRoleRequest contains the data for updating a user's role.
type UpdateRoleRequest struct {
	UserID    uuid.UUID
	TenantID  uuid.UUID
	NewRole   domain.Role
	UpdatedBy uuid.UUID
	IPAddress string
}

// UpdateRole changes a user's role in a tenant.
func (s *UserService) UpdateRole(ctx context.Context, req UpdateRoleRequest, callerRole domain.Role) error {
	// Check if caller can assign the new role
	if !callerRole.CanAssign(req.NewRole) {
		return domain.ErrCannotAssignRole
	}

	// Get current role assignment
	roleAssignment, err := s.roleRepo.FindByUserAndTenant(ctx, req.UserID, req.TenantID)
	if err != nil {
		return err
	}

	// Check if caller can manage the current role
	if !callerRole.CanManage(roleAssignment.Role) {
		return domain.ErrCannotManageRole
	}

	oldRole := roleAssignment.Role
	roleAssignment.Role = req.NewRole

	if err := s.roleRepo.Update(ctx, roleAssignment); err != nil {
		return err
	}

	// Log role change
	s.logEvent(ctx, domain.EventRoleChanged, &req.UserID, &req.TenantID, req.IPAddress, "", map[string]interface{}{
		"old_role":   oldRole,
		"new_role":   req.NewRole,
		"updated_by": req.UpdatedBy,
	})

	return nil
}

// List retrieves users in a tenant with pagination.
func (s *UserService) List(ctx context.Context, tenantID uuid.UUID, page, limit int) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.userRepo.ListByTenant(ctx, tenantID, offset, limit)
}

// ChangePasswordRequest contains the data for changing a password.
type ChangePasswordRequest struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
	IPAddress       string
}

// ChangePassword changes a user's password.
func (s *UserService) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	// Verify current password
	match, err := s.passwordSvc.Verify(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return err
	}
	if !match {
		return domain.ErrPasswordIncorrect
	}

	// Validate new password
	if err := s.passwordSvc.ValidatePassword(req.NewPassword); err != nil {
		return domain.ErrPasswordWeak
	}

	// Hash new password
	newHash, err := s.passwordSvc.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	user.MustResetPwd = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Revoke all sessions per FR-014
	if err := s.sessionRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		return err
	}

	// Log password change
	s.logEvent(ctx, domain.EventPasswordChanged, &user.ID, nil, req.IPAddress, "", nil)

	return nil
}

// RequestPasswordReset initiates a password reset flow.
func (s *UserService) RequestPasswordReset(ctx context.Context, email, ipAddress string) error {
	// Rate limit by email
	if s.resetRateLimiter != nil {
		allowed, err := s.resetRateLimiter.Allow(ctx, email)
		if err != nil {
			return err
		}
		if !allowed {
			return domain.ErrRateLimitExceeded
		}
	}

	// Find user - don't reveal if email exists
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Log attempt but don't reveal email doesn't exist
			s.logEvent(ctx, domain.EventPasswordResetRequested, nil, nil, ipAddress, "", map[string]interface{}{
				"email": email,
				"found": false,
			})
			return nil // Success response to prevent email enumeration
		}
		return err
	}

	// Generate reset token
	plainToken, tokenHash, err := s.passwordSvc.GenerateResetToken()
	if err != nil {
		return err
	}

	// Store token (1 hour expiry)
	resetToken := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := s.passwordReset.Create(ctx, resetToken); err != nil {
		return err
	}

	// Log reset request
	s.logEvent(ctx, domain.EventPasswordResetRequested, &user.ID, nil, ipAddress, "", map[string]interface{}{
		"found": true,
	})

	// TODO: Send email with plainToken
	// For now, we just store the token. Email sending will be added later.
	_ = plainToken

	return nil
}

// CompletePasswordReset completes the password reset flow.
func (s *UserService) CompletePasswordReset(ctx context.Context, token, newPassword, ipAddress string) error {
	// Validate new password first
	if err := s.passwordSvc.ValidatePassword(newPassword); err != nil {
		return domain.ErrPasswordWeak
	}

	// Hash the provided token
	tokenHash := s.passwordSvc.HashResetToken(token)

	// Find token
	resetToken, err := s.passwordReset.FindByToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	// Check if token is valid
	if resetToken.IsUsed() {
		return domain.ErrPasswordResetUsed
	}
	if resetToken.IsExpired() {
		return domain.ErrPasswordResetExpired
	}

	// Hash new password
	newHash, err := s.passwordSvc.Hash(newPassword)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, resetToken.UserID)
	if err != nil {
		return err
	}

	// Update password
	user.PasswordHash = newHash
	user.MustResetPwd = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Mark token as used
	if err := s.passwordReset.MarkUsed(ctx, resetToken.ID); err != nil {
		return err
	}

	// Revoke all sessions per FR-014
	if err := s.sessionRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		return err
	}

	// Log password reset completion
	s.logEvent(ctx, domain.EventPasswordResetCompleted, &user.ID, nil, ipAddress, "", nil)

	return nil
}

// logEvent logs an authentication event.
func (s *UserService) logEvent(ctx context.Context, eventType domain.AuthEventType, userID, tenantID *uuid.UUID, ipAddress, userAgent string, metadata map[string]interface{}) {
	event := domain.NewAuthEvent(eventType, userID, tenantID, ipAddress, userAgent)
	if metadata != nil {
		event.Metadata = metadata
	}
	_ = s.eventRepo.Create(ctx, event)
}
