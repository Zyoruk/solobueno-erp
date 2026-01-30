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
}
