package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing.
// Note: We manually create tables with SQLite-compatible schema instead of using
// AutoMigrate, which would try to use PostgreSQL-specific syntax from GORM tags.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create tables manually with SQLite-compatible schema
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	// SQLite-compatible schema
	schema := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			first_name TEXT,
			last_name TEXT,
			is_active INTEGER DEFAULT 1,
			must_reset_pwd INTEGER DEFAULT 0,
			failed_login_count INTEGER DEFAULT 0,
			locked_until DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS tenants (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			is_active INTEGER DEFAULT 1,
			settings TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS user_tenant_roles (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			tenant_id TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, tenant_id)
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			tenant_id TEXT NOT NULL,
			refresh_token TEXT UNIQUE NOT NULL,
			device_info TEXT,
			ip_address TEXT,
			expires_at DATETIME NOT NULL,
			revoked_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			used_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS auth_events (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			tenant_id TEXT,
			event_type TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err = sqlDB.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create test tables: %v", err)
	}

	return db
}

// ============ User Repository Tests ============

func TestGormUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "Test",
		LastName:     "User",
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify user was created
	found, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Email != user.Email {
		t.Errorf("Email = %q, want %q", found.Email, user.Email)
	}
}

func TestGormUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "findme@example.com",
		PasswordHash: "hashed",
		IsActive:     true,
	}
	repo.Create(ctx, user)

	found, err := repo.FindByEmail(ctx, "findme@example.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found.ID != user.ID {
		t.Error("Wrong user returned")
	}

	// Test not found
	_, err = repo.FindByEmail(ctx, "notfound@example.com")
	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestGormUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:        uuid.New(),
		Email:     "update@example.com",
		FirstName: "Before",
		IsActive:  true,
	}
	repo.Create(ctx, user)

	user.FirstName = "After"
	err := repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, user.ID)
	if found.FirstName != "After" {
		t.Errorf("FirstName = %q, want %q", found.FirstName, "After")
	}
}

func TestGormUserRepository_ExistsByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:    uuid.New(),
		Email: "exists@example.com",
	}
	repo.Create(ctx, user)

	exists, err := repo.ExistsByEmail(ctx, "exists@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}
	if !exists {
		t.Error("Should return true for existing email")
	}

	exists, _ = repo.ExistsByEmail(ctx, "notexists@example.com")
	if exists {
		t.Error("Should return false for non-existing email")
	}
}

// ============ Session Repository Tests ============

func TestGormSessionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		TenantID:     uuid.New(),
		RefreshToken: "token_hash",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	err := repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.RefreshToken != session.RefreshToken {
		t.Error("Session not saved correctly")
	}
}

func TestGormSessionRepository_FindByToken(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		TenantID:     uuid.New(),
		RefreshToken: "unique_token_hash",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	repo.Create(ctx, session)

	found, err := repo.FindByToken(ctx, "unique_token_hash")
	if err != nil {
		t.Fatalf("FindByToken failed: %v", err)
	}
	if found.ID != session.ID {
		t.Error("Wrong session returned")
	}

	_, err = repo.FindByToken(ctx, "nonexistent")
	if err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestGormSessionRepository_Revoke(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		TenantID:     uuid.New(),
		RefreshToken: "revoke_test",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	repo.Create(ctx, session)

	err := repo.Revoke(ctx, session.ID)
	if err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, session.ID)
	if found.RevokedAt == nil {
		t.Error("Session should be revoked")
	}
}

func TestGormSessionRepository_RevokeAllForUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	tenantID := uuid.New()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session := &domain.Session{
			ID:           uuid.New(),
			UserID:       userID,
			TenantID:     tenantID,
			RefreshToken: uuid.New().String(),
			ExpiresAt:    time.Now().Add(time.Hour),
		}
		repo.Create(ctx, session)
	}

	err := repo.RevokeAllForUser(ctx, userID)
	if err != nil {
		t.Fatalf("RevokeAllForUser failed: %v", err)
	}

	count, _ := repo.CountActiveForUser(ctx, userID)
	if count != 0 {
		t.Errorf("Active sessions = %d, want 0", count)
	}
}

// ============ Tenant Repository Tests ============

func TestGormTenantRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormTenantRepository(db)
	ctx := context.Background()

	tenant := &domain.Tenant{
		ID:       uuid.New(),
		Name:     "Test Restaurant",
		Slug:     "test-restaurant",
		IsActive: true,
	}

	err := repo.Create(ctx, tenant)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != tenant.Name {
		t.Error("Tenant not saved correctly")
	}
}

func TestGormTenantRepository_FindBySlug(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormTenantRepository(db)
	ctx := context.Background()

	tenant := &domain.Tenant{
		ID:       uuid.New(),
		Name:     "Slug Test",
		Slug:     "slug-test",
		IsActive: true,
	}
	repo.Create(ctx, tenant)

	found, err := repo.FindBySlug(ctx, "slug-test")
	if err != nil {
		t.Fatalf("FindBySlug failed: %v", err)
	}
	if found.ID != tenant.ID {
		t.Error("Wrong tenant returned")
	}

	_, err = repo.FindBySlug(ctx, "nonexistent")
	if err != domain.ErrTenantNotFound {
		t.Errorf("Expected ErrTenantNotFound, got %v", err)
	}
}

// ============ Auth Event Repository Tests ============

func TestGormAuthEventRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormAuthEventRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	event := domain.NewAuthEvent(domain.EventLoginSuccess, &userID, nil, "127.0.0.1", "TestAgent")

	err := repo.Create(ctx, event)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	events, total, err := repo.FindByUser(ctx, userID, 0, 10)
	if err != nil {
		t.Fatalf("FindByUser failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Total = %d, want 1", total)
	}
	if len(events) != 1 {
		t.Errorf("len(events) = %d, want 1", len(events))
	}
}

// ============ Password Reset Repository Tests ============

func TestGormPasswordResetRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordResetRepository(db)
	ctx := context.Background()

	token := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "reset_token_hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err := repo.Create(ctx, token)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByToken(ctx, "reset_token_hash")
	if err != nil {
		t.Fatalf("FindByToken failed: %v", err)
	}
	if found.ID != token.ID {
		t.Error("Wrong token returned")
	}
}

func TestGormPasswordResetRepository_MarkUsed(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordResetRepository(db)
	ctx := context.Background()

	token := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "mark_used_test",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo.Create(ctx, token)

	err := repo.MarkUsed(ctx, token.ID)
	if err != nil {
		t.Fatalf("MarkUsed failed: %v", err)
	}

	found, _ := repo.FindByToken(ctx, "mark_used_test")
	if found.UsedAt == nil {
		t.Error("Token should be marked as used")
	}
}

// ============ UserTenantRole Repository Tests ============

func TestGormUserTenantRoleRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserTenantRoleRepository(db)
	ctx := context.Background()

	// Create user and tenant first
	userRepo := NewGormUserRepository(db)
	tenantRepo := NewGormTenantRepository(db)

	user := &domain.User{ID: uuid.New(), Email: "role@test.com"}
	tenant := &domain.Tenant{ID: uuid.New(), Name: "Role Test", Slug: "role-test"}
	userRepo.Create(ctx, user)
	tenantRepo.Create(ctx, tenant)

	role := &domain.UserTenantRole{
		ID:       uuid.New(),
		UserID:   user.ID,
		TenantID: tenant.ID,
		Role:     domain.RoleManager,
	}

	err := repo.Create(ctx, role)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByUserAndTenant(ctx, user.ID, tenant.ID)
	if err != nil {
		t.Fatalf("FindByUserAndTenant failed: %v", err)
	}
	if found.Role != domain.RoleManager {
		t.Errorf("Role = %v, want %v", found.Role, domain.RoleManager)
	}

	_, err = repo.FindByUserAndTenant(ctx, uuid.New(), tenant.ID)
	if err != domain.ErrUserNotInTenant {
		t.Errorf("Expected ErrUserNotInTenant, got %v", err)
	}
}

func TestGormUserTenantRoleRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserTenantRoleRepository(db)
	ctx := context.Background()

	userID, tenantID := uuid.New(), uuid.New()
	role := &domain.UserTenantRole{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleWaiter}
	repo.Create(ctx, role)

	role.Role = domain.RoleManager
	if err := repo.Update(ctx, role); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByUserAndTenant(ctx, userID, tenantID)
	if found.Role != domain.RoleManager {
		t.Errorf("Role = %v, want %v", found.Role, domain.RoleManager)
	}
}

func TestGormUserTenantRoleRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserTenantRoleRepository(db)
	ctx := context.Background()

	userID, tenantID := uuid.New(), uuid.New()
	role := &domain.UserTenantRole{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleWaiter}
	repo.Create(ctx, role)

	if err := repo.Delete(ctx, role.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := repo.FindByUserAndTenant(ctx, userID, tenantID); err != domain.ErrUserNotInTenant {
		t.Errorf("Expected ErrUserNotInTenant after delete, got %v", err)
	}
}

func TestGormUserTenantRoleRepository_DeleteByUserAndTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserTenantRoleRepository(db)
	ctx := context.Background()

	userID, tenantID := uuid.New(), uuid.New()
	role := &domain.UserTenantRole{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleWaiter}
	repo.Create(ctx, role)

	if err := repo.DeleteByUserAndTenant(ctx, userID, tenantID); err != nil {
		t.Fatalf("DeleteByUserAndTenant failed: %v", err)
	}

	if err := repo.DeleteByUserAndTenant(ctx, userID, tenantID); err != domain.ErrUserNotInTenant {
		t.Errorf("Expected ErrUserNotInTenant on repeat delete, got %v", err)
	}
}

func TestGormUserTenantRoleRepository_ListByUserAndListByTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserTenantRoleRepository(db)
	ctx := context.Background()

	userID, tenantID := uuid.New(), uuid.New()
	repo.Create(ctx, &domain.UserTenantRole{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleWaiter})
	repo.Create(ctx, &domain.UserTenantRole{ID: uuid.New(), UserID: userID, TenantID: uuid.New(), Role: domain.RoleCashier})

	byUser, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser failed: %v", err)
	}
	if len(byUser) != 2 {
		t.Errorf("ListByUser len = %d, want 2", len(byUser))
	}

	byTenant, err := repo.ListByTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if len(byTenant) != 1 {
		t.Errorf("ListByTenant len = %d, want 1", len(byTenant))
	}
}

// ============ Additional User Repository Coverage ============

func TestGormUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)

	_, err := repo.FindByID(context.Background(), uuid.New())
	if err != domain.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestGormUserRepository_WithTenants(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	user := &domain.User{ID: uuid.New(), Email: "withtenants@example.com", IsActive: true}
	repo.Create(ctx, user)

	byID, err := repo.FindByIDWithTenants(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByIDWithTenants failed: %v", err)
	}
	if byID.ID != user.ID {
		t.Error("Wrong user returned by FindByIDWithTenants")
	}

	byEmail, err := repo.FindByEmailWithTenants(ctx, user.Email)
	if err != nil {
		t.Fatalf("FindByEmailWithTenants failed: %v", err)
	}
	if byEmail.ID != user.ID {
		t.Error("Wrong user returned by FindByEmailWithTenants")
	}
}

func TestGormUserRepository_ListByTenant(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewGormUserRepository(db)
	tenantRepo := NewGormTenantRepository(db)
	roleRepo := NewGormUserTenantRoleRepository(db)
	ctx := context.Background()

	tenant := &domain.Tenant{ID: uuid.New(), Name: "List Tenant", Slug: "list-tenant"}
	tenantRepo.Create(ctx, tenant)

	for i := 0; i < 2; i++ {
		user := &domain.User{ID: uuid.New(), Email: uuid.New().String() + "@example.com", IsActive: true}
		userRepo.Create(ctx, user)
		roleRepo.Create(ctx, &domain.UserTenantRole{ID: uuid.New(), UserID: user.ID, TenantID: tenant.ID, Role: domain.RoleWaiter})
	}

	users, total, err := userRepo.ListByTenant(ctx, tenant.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListByTenant failed: %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Errorf("ListByTenant = %d/%d, want 2/2", len(users), total)
	}
}

// ============ Additional Session Repository Coverage ============

func TestGormSessionRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)

	_, err := repo.FindByID(context.Background(), uuid.New())
	if err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestGormSessionRepository_RevokeByToken(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	session := &domain.Session{ID: uuid.New(), UserID: uuid.New(), TenantID: uuid.New(), RefreshToken: "revoke_by_token", ExpiresAt: time.Now().Add(time.Hour)}
	repo.Create(ctx, session)

	if err := repo.RevokeByToken(ctx, "revoke_by_token"); err != nil {
		t.Fatalf("RevokeByToken failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, session.ID)
	if found.RevokedAt == nil {
		t.Error("Session should be revoked")
	}

	if err := repo.RevokeByToken(ctx, "nonexistent"); err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestGormSessionRepository_RevokeAllForUserInTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	userID, tenantA, tenantB := uuid.New(), uuid.New(), uuid.New()
	repo.Create(ctx, &domain.Session{ID: uuid.New(), UserID: userID, TenantID: tenantA, RefreshToken: uuid.New().String(), ExpiresAt: time.Now().Add(time.Hour)})
	repo.Create(ctx, &domain.Session{ID: uuid.New(), UserID: userID, TenantID: tenantB, RefreshToken: uuid.New().String(), ExpiresAt: time.Now().Add(time.Hour)})

	if err := repo.RevokeAllForUserInTenant(ctx, userID, tenantA); err != nil {
		t.Fatalf("RevokeAllForUserInTenant failed: %v", err)
	}

	countA, _ := repo.CountActiveForUser(ctx, userID)
	if countA != 1 {
		t.Errorf("Active sessions across tenants = %d, want 1 (tenantB still active)", countA)
	}
}

func TestGormSessionRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.Session{ID: uuid.New(), UserID: uuid.New(), TenantID: uuid.New(), RefreshToken: uuid.New().String(), ExpiresAt: time.Now().Add(-time.Hour)})
	repo.Create(ctx, &domain.Session{ID: uuid.New(), UserID: uuid.New(), TenantID: uuid.New(), RefreshToken: uuid.New().String(), ExpiresAt: time.Now().Add(time.Hour)})

	deleted, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("DeleteExpired = %d, want 1", deleted)
	}
}

// ============ Additional Tenant Repository Coverage ============

func TestGormTenantRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormTenantRepository(db)

	_, err := repo.FindByID(context.Background(), uuid.New())
	if err != domain.ErrTenantNotFound {
		t.Errorf("Expected ErrTenantNotFound, got %v", err)
	}
}

func TestGormTenantRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormTenantRepository(db)
	ctx := context.Background()

	tenant := &domain.Tenant{ID: uuid.New(), Name: "Before", Slug: "update-tenant", IsActive: true}
	repo.Create(ctx, tenant)

	tenant.Name = "After"
	if err := repo.Update(ctx, tenant); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, tenant.ID)
	if found.Name != "After" {
		t.Errorf("Name = %q, want %q", found.Name, "After")
	}
}

func TestGormTenantRepository_ExistsBySlug(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormTenantRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.Tenant{ID: uuid.New(), Name: "Exists", Slug: "exists-slug"})

	exists, err := repo.ExistsBySlug(ctx, "exists-slug")
	if err != nil {
		t.Fatalf("ExistsBySlug failed: %v", err)
	}
	if !exists {
		t.Error("Should return true for existing slug")
	}

	exists, _ = repo.ExistsBySlug(ctx, "missing-slug")
	if exists {
		t.Error("Should return false for missing slug")
	}
}

// ============ Additional Auth Event Repository Coverage ============

func TestGormAuthEventRepository_FindByTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormAuthEventRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	repo.Create(ctx, domain.NewAuthEvent(domain.EventLoginSuccess, nil, &tenantID, "127.0.0.1", "TestAgent"))

	events, total, err := repo.FindByTenant(ctx, tenantID, 0, 10)
	if err != nil {
		t.Fatalf("FindByTenant failed: %v", err)
	}
	if total != 1 || len(events) != 1 {
		t.Errorf("FindByTenant = %d/%d, want 1/1", len(events), total)
	}
}

func TestGormAuthEventRepository_FindByType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormAuthEventRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	repo.Create(ctx, domain.NewAuthEvent(domain.EventLoginFailed, &userID, nil, "127.0.0.1", "TestAgent"))

	events, total, err := repo.FindByType(ctx, domain.EventLoginFailed, 0, 10)
	if err != nil {
		t.Fatalf("FindByType failed: %v", err)
	}
	if total != 1 || len(events) != 1 {
		t.Errorf("FindByType = %d/%d, want 1/1", len(events), total)
	}
}

func TestGormAuthEventRepository_FindByUserAndType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormAuthEventRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	repo.Create(ctx, domain.NewAuthEvent(domain.EventLoginSuccess, &userID, nil, "127.0.0.1", "TestAgent"))

	events, err := repo.FindByUserAndType(ctx, userID, domain.EventLoginSuccess, time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("FindByUserAndType failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("len(events) = %d, want 1", len(events))
	}
}

func TestGormAuthEventRepository_CountRecentByIP(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormAuthEventRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		repo.Create(ctx, domain.NewAuthEvent(domain.EventLoginFailed, nil, nil, "10.0.0.1", "TestAgent"))
	}

	count, err := repo.CountRecentByIP(ctx, "10.0.0.1", domain.EventLoginFailed, time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("CountRecentByIP failed: %v", err)
	}
	if count != 3 {
		t.Errorf("CountRecentByIP = %d, want 3", count)
	}
}

func TestGormAuthEventRepository_DeleteOlderThan(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormAuthEventRepository(db)
	ctx := context.Background()

	event := domain.NewAuthEvent(domain.EventLogout, nil, nil, "127.0.0.1", "TestAgent")
	repo.Create(ctx, event)

	deleted, err := repo.DeleteOlderThan(ctx, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("DeleteOlderThan failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("DeleteOlderThan = %d, want 1", deleted)
	}
}

// ============ Additional Password Reset Repository Coverage ============

func TestGormPasswordResetRepository_FindByToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordResetRepository(db)

	_, err := repo.FindByToken(context.Background(), "missing")
	if err != domain.ErrPasswordResetInvalid {
		t.Errorf("Expected ErrPasswordResetInvalid, got %v", err)
	}
}

func TestGormPasswordResetRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordResetRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &domain.PasswordResetToken{ID: uuid.New(), UserID: uuid.New(), TokenHash: "expired", ExpiresAt: time.Now().Add(-time.Hour)})
	repo.Create(ctx, &domain.PasswordResetToken{ID: uuid.New(), UserID: uuid.New(), TokenHash: "valid", ExpiresAt: time.Now().Add(time.Hour)})

	deleted, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("DeleteExpired = %d, want 1", deleted)
	}
}

func TestGormPasswordResetRepository_DeleteForUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordResetRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	repo.Create(ctx, &domain.PasswordResetToken{ID: uuid.New(), UserID: userID, TokenHash: "for_user", ExpiresAt: time.Now().Add(time.Hour)})

	if err := repo.DeleteForUser(ctx, userID); err != nil {
		t.Fatalf("DeleteForUser failed: %v", err)
	}

	if _, err := repo.FindByToken(ctx, "for_user"); err != domain.ErrPasswordResetInvalid {
		t.Errorf("Expected token removed, got %v", err)
	}
}

// ============ Auto-generated ID coverage (zero-value ID on Create) ============

func TestGormRepositories_CreateGeneratesID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	user := &domain.User{Email: "autoid-user@example.com"}
	if err := NewGormUserRepository(db).Create(ctx, user); err != nil {
		t.Fatalf("user Create failed: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Error("user Create should generate an ID")
	}

	tenant := &domain.Tenant{Name: "AutoID Tenant", Slug: "autoid-tenant"}
	if err := NewGormTenantRepository(db).Create(ctx, tenant); err != nil {
		t.Fatalf("tenant Create failed: %v", err)
	}
	if tenant.ID == uuid.Nil {
		t.Error("tenant Create should generate an ID")
	}

	role := &domain.UserTenantRole{UserID: user.ID, TenantID: tenant.ID, Role: domain.RoleWaiter}
	if err := NewGormUserTenantRoleRepository(db).Create(ctx, role); err != nil {
		t.Fatalf("role Create failed: %v", err)
	}
	if role.ID == uuid.Nil {
		t.Error("role Create should generate an ID")
	}

	session := &domain.Session{UserID: user.ID, TenantID: tenant.ID, RefreshToken: "autoid-session", ExpiresAt: time.Now().Add(time.Hour)}
	if err := NewGormSessionRepository(db).Create(ctx, session); err != nil {
		t.Fatalf("session Create failed: %v", err)
	}
	if session.ID == uuid.Nil {
		t.Error("session Create should generate an ID")
	}

	token := &domain.PasswordResetToken{UserID: user.ID, TokenHash: "autoid-token", ExpiresAt: time.Now().Add(time.Hour)}
	if err := NewGormPasswordResetRepository(db).Create(ctx, token); err != nil {
		t.Fatalf("token Create failed: %v", err)
	}
	if token.ID == uuid.Nil {
		t.Error("token Create should generate an ID")
	}

	event := domain.NewAuthEvent(domain.EventLoginSuccess, &user.ID, nil, "127.0.0.1", "TestAgent")
	event.ID = uuid.Nil
	if err := NewGormAuthEventRepository(db).Create(ctx, event); err != nil {
		t.Fatalf("event Create failed: %v", err)
	}
	if event.ID == uuid.Nil {
		t.Error("event Create should generate an ID")
	}
}

// ============ DB-error propagation coverage ============
// Closing the underlying connection forces GORM to return a non-NotFound
// error, exercising the generic error-passthrough branches.

func TestGormRepositories_PropagateDBErrors(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.Close()
	ctx := context.Background()

	if _, err := NewGormUserRepository(db).FindByID(ctx, uuid.New()); err == nil || err == domain.ErrUserNotFound {
		t.Errorf("expected non-NotFound DB error, got %v", err)
	}
	if _, err := NewGormSessionRepository(db).FindByID(ctx, uuid.New()); err == nil || err == domain.ErrSessionNotFound {
		t.Errorf("expected non-NotFound DB error, got %v", err)
	}
	if _, err := NewGormTenantRepository(db).FindByID(ctx, uuid.New()); err == nil || err == domain.ErrTenantNotFound {
		t.Errorf("expected non-NotFound DB error, got %v", err)
	}
	if _, err := NewGormPasswordResetRepository(db).FindByToken(ctx, "x"); err == nil || err == domain.ErrPasswordResetInvalid {
		t.Errorf("expected non-NotFound DB error, got %v", err)
	}
	if _, err := NewGormUserTenantRoleRepository(db).FindByUserAndTenant(ctx, uuid.New(), uuid.New()); err == nil || err == domain.ErrUserNotInTenant {
		t.Errorf("expected non-NotFound DB error, got %v", err)
	}
}

func TestGormPasswordResetRepository_CountRecentForUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordResetRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	repo.Create(ctx, &domain.PasswordResetToken{ID: uuid.New(), UserID: userID, TokenHash: "recent1", ExpiresAt: time.Now().Add(time.Hour)})
	repo.Create(ctx, &domain.PasswordResetToken{ID: uuid.New(), UserID: userID, TokenHash: "recent2", ExpiresAt: time.Now().Add(time.Hour)})

	count, err := repo.CountRecentForUser(ctx, userID, time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("CountRecentForUser failed: %v", err)
	}
	if count != 2 {
		t.Errorf("CountRecentForUser = %d, want 2", count)
	}
}
