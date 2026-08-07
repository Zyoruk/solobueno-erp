package auth

import (
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/gorm"
)

// AutoMigrate runs GORM auto-migration for all auth domain models.
// This is intended for development use. For production, use explicit SQL migrations.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.Tenant{},
		&domain.UserTenantRole{},
		&domain.Session{},
		&domain.PasswordResetToken{},
		&domain.AuthEvent{},
	)
}

// DropAll drops all auth-related tables.
// WARNING: This is destructive and should only be used in development/testing.
func DropAll(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&domain.AuthEvent{},
		&domain.PasswordResetToken{},
		&domain.Session{},
		&domain.UserTenantRole{},
		&domain.Tenant{},
		&domain.User{},
	)
}
