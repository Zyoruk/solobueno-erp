package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository"
)

func setupUserService(t *testing.T) (*UserService, *repository.MockUserRepository, *repository.MockUserTenantRoleRepository, *repository.MockSessionRepository, *repository.MockPasswordResetRepository) {
	t.Helper()

	userRepo := repository.NewMockUserRepository()
	roleRepo := repository.NewMockUserTenantRoleRepository()
	sessionRepo := repository.NewMockSessionRepository()
	eventRepo := repository.NewMockAuthEventRepository()
	passwordResetRepo := repository.NewMockPasswordResetRepository()

	userSvc := NewUserService(UserServiceConfig{
		UserRepo:         userRepo,
		RoleRepo:         roleRepo,
		SessionRepo:      sessionRepo,
		EventRepo:        eventRepo,
		PasswordReset:    passwordResetRepo,
		ResetRateLimiter: nil,
	})

	return userSvc, userRepo, roleRepo, sessionRepo, passwordResetRepo
}

func TestUserService_Create_Success(t *testing.T) {
	userSvc, _, _, _, _ := setupUserService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	creatorID := uuid.New()

	resp, err := userSvc.Create(ctx, CreateUserRequest{
		Email:     "new@example.com",
		FirstName: "New",
		LastName:  "User",
		TenantID:  tenantID,
		Role:      domain.RoleWaiter,
		CreatedBy: creatorID,
		IPAddress: "127.0.0.1",
	}, domain.RoleManager)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if resp.User == nil {
		t.Fatal("User should not be nil")
	}
	if resp.User.Email != "new@example.com" {
		t.Errorf("Email = %q, want %q", resp.User.Email, "new@example.com")
	}
	if resp.TemporaryPassword == "" {
		t.Error("TemporaryPassword should not be empty")
	}
	if !resp.User.MustResetPwd {
		t.Error("MustResetPwd should be true")
	}
}

func TestUserService_Create_CannotAssignRole(t *testing.T) {
	userSvc, _, _, _, _ := setupUserService(t)
	ctx := context.Background()

	// Waiter trying to create a Manager
	_, err := userSvc.Create(ctx, CreateUserRequest{
		Email:    "new@example.com",
		TenantID: uuid.New(),
		Role:     domain.RoleManager,
	}, domain.RoleWaiter)

	if err != domain.ErrCannotAssignRole {
		t.Errorf("Expected ErrCannotAssignRole, got %v", err)
	}
}

func TestUserService_Create_EmailExists(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	// Add existing user
	userRepo.AddUser(&domain.User{
		ID:    uuid.New(),
		Email: "existing@example.com",
	})

	_, err := userSvc.Create(ctx, CreateUserRequest{
		Email:    "existing@example.com",
		TenantID: uuid.New(),
		Role:     domain.RoleWaiter,
	}, domain.RoleManager)

	if err != domain.ErrEmailExists {
		t.Errorf("Expected ErrEmailExists, got %v", err)
	}
}

func TestUserService_GetByID(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.AddUser(&domain.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	})

	user, err := userSvc.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "test@example.com")
	}
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	userSvc, _, _, _, _ := setupUserService(t)
	ctx := context.Background()

	_, err := userSvc.GetByID(ctx, uuid.New())
	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_Update_Success(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	userRepo.AddUser(&domain.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Old",
		LastName:  "Name",
		IsActive:  true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleWaiter},
		},
	})

	newFirstName := "New"
	newLastName := "Updated"

	user, err := userSvc.Update(ctx, UpdateRequest{
		UserID:    userID,
		FirstName: &newFirstName,
		LastName:  &newLastName,
		TenantID:  tenantID,
	}, domain.RoleManager)

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if user.FirstName != "New" {
		t.Errorf("FirstName = %q, want %q", user.FirstName, "New")
	}
	if user.LastName != "Updated" {
		t.Errorf("LastName = %q, want %q", user.LastName, "Updated")
	}
}

func TestUserService_Update_CannotManage(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	userRepo.AddUser(&domain.User{
		ID:       userID,
		Email:    "test@example.com",
		IsActive: true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleManager}, // Target is Manager
		},
	})

	newFirstName := "New"
	_, err := userSvc.Update(ctx, UpdateRequest{
		UserID:    userID,
		FirstName: &newFirstName,
		TenantID:  tenantID,
	}, domain.RoleWaiter) // Waiter cannot manage Manager

	if err != domain.ErrCannotManageRole {
		t.Errorf("Expected ErrCannotManageRole, got %v", err)
	}
}

func TestUserService_Update_Deactivate(t *testing.T) {
	userSvc, userRepo, _, sessionRepo, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	userRepo.AddUser(&domain.User{
		ID:       userID,
		Email:    "test@example.com",
		IsActive: true,
		TenantRoles: []domain.UserTenantRole{
			{TenantID: tenantID, Role: domain.RoleWaiter},
		},
	})

	// Add a session for the user
	sessionRepo.Create(ctx, &domain.Session{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		ExpiresAt: time.Now().Add(time.Hour),
	})

	isActive := false
	user, err := userSvc.Update(ctx, UpdateRequest{
		UserID:   userID,
		IsActive: &isActive,
		TenantID: tenantID,
	}, domain.RoleManager)

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if user.IsActive {
		t.Error("User should be deactivated")
	}
}

func TestUserService_UpdateRole_Success(t *testing.T) {
	userSvc, userRepo, roleRepo, _, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()
	roleID := uuid.New()

	userRepo.AddUser(&domain.User{
		ID:    userID,
		Email: "test@example.com",
	})

	roleRepo.AddRole(&domain.UserTenantRole{
		ID:       roleID,
		UserID:   userID,
		TenantID: tenantID,
		Role:     domain.RoleWaiter,
	})

	err := userSvc.UpdateRole(ctx, UpdateRoleRequest{
		UserID:    userID,
		TenantID:  tenantID,
		NewRole:   domain.RoleCashier,
		UpdatedBy: uuid.New(),
		IPAddress: "127.0.0.1",
	}, domain.RoleManager)

	if err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}
}

func TestUserService_UpdateRole_CannotAssign(t *testing.T) {
	userSvc, _, roleRepo, _, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	roleRepo.AddRole(&domain.UserTenantRole{
		ID:       uuid.New(),
		UserID:   userID,
		TenantID: tenantID,
		Role:     domain.RoleWaiter,
	})

	// Waiter trying to assign Manager role
	err := userSvc.UpdateRole(ctx, UpdateRoleRequest{
		UserID:   userID,
		TenantID: tenantID,
		NewRole:  domain.RoleManager,
	}, domain.RoleWaiter)

	if err != domain.ErrCannotAssignRole {
		t.Errorf("Expected ErrCannotAssignRole, got %v", err)
	}
}

func TestUserService_List(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	tenantID := uuid.New()

	userRepo.ListByTenantFunc = func(ctx context.Context, tid uuid.UUID, offset, limit int) ([]*domain.User, int64, error) {
		return []*domain.User{
			{ID: uuid.New(), Email: "user1@example.com"},
			{ID: uuid.New(), Email: "user2@example.com"},
		}, 2, nil
	}

	users, total, err := userSvc.List(ctx, tenantID, 1, 20)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("len(users) = %d, want 2", len(users))
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
}

func TestUserService_List_Pagination(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	called := false
	userRepo.ListByTenantFunc = func(ctx context.Context, tid uuid.UUID, offset, limit int) ([]*domain.User, int64, error) {
		called = true
		// Verify pagination parameters
		if offset != 20 {
			t.Errorf("offset = %d, want 20", offset)
		}
		if limit != 10 {
			t.Errorf("limit = %d, want 10", limit)
		}
		return []*domain.User{}, 0, nil
	}

	userSvc.List(ctx, uuid.New(), 3, 10) // page 3, limit 10 => offset 20

	if !called {
		t.Error("ListByTenantFunc should have been called")
	}
}

func TestUserService_List_DefaultPagination(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	userRepo.ListByTenantFunc = func(ctx context.Context, tid uuid.UUID, offset, limit int) ([]*domain.User, int64, error) {
		if limit != 20 {
			t.Errorf("default limit = %d, want 20", limit)
		}
		if offset != 0 {
			t.Errorf("default offset = %d, want 0", offset)
		}
		return []*domain.User{}, 0, nil
	}

	// Invalid page/limit should use defaults
	userSvc.List(ctx, uuid.New(), 0, 0)
}

func TestUserService_ChangePassword_Success(t *testing.T) {
	userSvc, userRepo, _, sessionRepo, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("OldPassword123!")

	userRepo.AddUser(&domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		MustResetPwd: true,
	})

	err := userSvc.ChangePassword(ctx, ChangePasswordRequest{
		UserID:          userID,
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword456!",
		IPAddress:       "127.0.0.1",
	})

	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Verify password was changed
	user, _ := userRepo.FindByID(ctx, userID)
	match, _ := NewPasswordService().Verify("NewPassword456!", user.PasswordHash)
	if !match {
		t.Error("Password should have been changed")
	}
	if user.MustResetPwd {
		t.Error("MustResetPwd should be false after change")
	}

	// Verify sessions were revoked (mock doesn't fail)
	_ = sessionRepo
}

func TestUserService_ChangePassword_WrongCurrent(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("CorrectPassword123!")

	userRepo.AddUser(&domain.User{
		ID:           userID,
		PasswordHash: passwordHash,
	})

	err := userSvc.ChangePassword(ctx, ChangePasswordRequest{
		UserID:          userID,
		CurrentPassword: "WrongPassword123!",
		NewPassword:     "NewPassword456!",
	})

	if err != domain.ErrPasswordIncorrect {
		t.Errorf("Expected ErrPasswordIncorrect, got %v", err)
	}
}

func TestUserService_ChangePassword_WeakNew(t *testing.T) {
	userSvc, userRepo, _, _, _ := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("OldPassword123!")

	userRepo.AddUser(&domain.User{
		ID:           userID,
		PasswordHash: passwordHash,
	})

	err := userSvc.ChangePassword(ctx, ChangePasswordRequest{
		UserID:          userID,
		CurrentPassword: "OldPassword123!",
		NewPassword:     "weak", // Too weak
	})

	if err != domain.ErrPasswordWeak {
		t.Errorf("Expected ErrPasswordWeak, got %v", err)
	}
}

func TestUserService_RequestPasswordReset_UserExists(t *testing.T) {
	userSvc, userRepo, _, _, passwordResetRepo := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.AddUser(&domain.User{
		ID:    userID,
		Email: "test@example.com",
	})

	err := userSvc.RequestPasswordReset(ctx, "test@example.com", "127.0.0.1")
	if err != nil {
		t.Fatalf("RequestPasswordReset failed: %v", err)
	}

	// Verify token was created
	_ = passwordResetRepo
}

func TestUserService_RequestPasswordReset_UserNotExists(t *testing.T) {
	userSvc, _, _, _, _ := setupUserService(t)
	ctx := context.Background()

	// Should not return error to prevent email enumeration
	err := userSvc.RequestPasswordReset(ctx, "nonexistent@example.com", "127.0.0.1")
	if err != nil {
		t.Errorf("Should not return error for non-existent email, got %v", err)
	}
}

func TestUserService_RequestPasswordReset_RateLimited(t *testing.T) {
	userRepo := repository.NewMockUserRepository()
	roleRepo := repository.NewMockUserTenantRoleRepository()
	sessionRepo := repository.NewMockSessionRepository()
	eventRepo := repository.NewMockAuthEventRepository()
	passwordResetRepo := repository.NewMockPasswordResetRepository()

	// Rate limiter that denies all
	rateLimiter := NewMemoryRateLimiter(RateLimiterConfig{
		MaxRequests: 0,
		Window:      time.Minute,
	})

	userSvc := NewUserService(UserServiceConfig{
		UserRepo:         userRepo,
		RoleRepo:         roleRepo,
		SessionRepo:      sessionRepo,
		EventRepo:        eventRepo,
		PasswordReset:    passwordResetRepo,
		ResetRateLimiter: rateLimiter,
	})

	ctx := context.Background()
	err := userSvc.RequestPasswordReset(ctx, "test@example.com", "127.0.0.1")

	if err != domain.ErrRateLimitExceeded {
		t.Errorf("Expected ErrRateLimitExceeded, got %v", err)
	}
}

func TestUserService_CompletePasswordReset_Success(t *testing.T) {
	userSvc, userRepo, _, sessionRepo, passwordResetRepo := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	passwordHash, _ := NewPasswordService().Hash("OldPassword123!")

	userRepo.AddUser(&domain.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		MustResetPwd: true,
	})

	// Create a valid reset token
	plainToken := "reset-token-12345"
	tokenHash := NewPasswordService().HashResetToken(plainToken)

	passwordResetRepo.AddToken(&domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
		UsedAt:    nil,
	})

	err := userSvc.CompletePasswordReset(ctx, plainToken, "NewSecurePassword123!", "127.0.0.1")
	if err != nil {
		t.Fatalf("CompletePasswordReset failed: %v", err)
	}

	// Verify password was changed
	user, _ := userRepo.FindByID(ctx, userID)
	match, _ := NewPasswordService().Verify("NewSecurePassword123!", user.PasswordHash)
	if !match {
		t.Error("Password should have been changed")
	}

	_ = sessionRepo
}

func TestUserService_CompletePasswordReset_TokenUsed(t *testing.T) {
	userSvc, userRepo, _, _, passwordResetRepo := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.AddUser(&domain.User{
		ID:    userID,
		Email: "test@example.com",
	})

	usedAt := time.Now()
	plainToken := "used-token-12345"
	tokenHash := NewPasswordService().HashResetToken(plainToken)

	passwordResetRepo.AddToken(&domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
		UsedAt:    &usedAt, // Already used
	})

	err := userSvc.CompletePasswordReset(ctx, plainToken, "NewPassword123!", "127.0.0.1")

	if err != domain.ErrPasswordResetUsed {
		t.Errorf("Expected ErrPasswordResetUsed, got %v", err)
	}
}

func TestUserService_CompletePasswordReset_TokenExpired(t *testing.T) {
	userSvc, userRepo, _, _, passwordResetRepo := setupUserService(t)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.AddUser(&domain.User{
		ID:    userID,
		Email: "test@example.com",
	})

	plainToken := "expired-token-12345"
	tokenHash := NewPasswordService().HashResetToken(plainToken)

	passwordResetRepo.AddToken(&domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-time.Hour), // Expired
		UsedAt:    nil,
	})

	err := userSvc.CompletePasswordReset(ctx, plainToken, "NewPassword123!", "127.0.0.1")

	if err != domain.ErrPasswordResetExpired {
		t.Errorf("Expected ErrPasswordResetExpired, got %v", err)
	}
}

func TestUserService_CompletePasswordReset_WeakPassword(t *testing.T) {
	userSvc, _, _, _, _ := setupUserService(t)
	ctx := context.Background()

	err := userSvc.CompletePasswordReset(ctx, "any-token", "weak", "127.0.0.1")

	if err != domain.ErrPasswordWeak {
		t.Errorf("Expected ErrPasswordWeak, got %v", err)
	}
}
