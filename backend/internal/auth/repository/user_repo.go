// Package repository provides data access interfaces and implementations for the auth module.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	// FindByID retrieves a user by their ID.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// FindByEmail retrieves a user by their email address.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// FindByEmailWithTenants retrieves a user with their tenant roles preloaded.
	FindByEmailWithTenants(ctx context.Context, email string) (*domain.User, error)

	// FindByIDWithTenants retrieves a user with their tenant roles preloaded.
	FindByIDWithTenants(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// Create creates a new user.
	Create(ctx context.Context, user *domain.User) error

	// Update updates an existing user.
	Update(ctx context.Context, user *domain.User) error

	// ListByTenant retrieves all users belonging to a tenant.
	ListByTenant(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*domain.User, int64, error)

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// GormUserRepository is a GORM implementation of UserRepository.
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GormUserRepository.
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindByID retrieves a user by their ID.
func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail retrieves a user by their email address.
func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmailWithTenants retrieves a user with their tenant roles preloaded.
func (r *GormUserRepository) FindByEmailWithTenants(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).
		Preload("TenantRoles").
		Preload("TenantRoles.Tenant").
		First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByIDWithTenants retrieves a user with their tenant roles preloaded.
func (r *GormUserRepository) FindByIDWithTenants(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).
		Preload("TenantRoles").
		Preload("TenantRoles.Tenant").
		First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Create creates a new user.
func (r *GormUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(user).Error
}

// Update updates an existing user.
func (r *GormUserRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// ListByTenant retrieves all users belonging to a tenant with pagination.
func (r *GormUserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Joins("JOIN user_tenant_roles ON user_tenant_roles.user_id = users.id").
		Where("user_tenant_roles.tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_tenant_roles ON user_tenant_roles.user_id = users.id").
		Where("user_tenant_roles.tenant_id = ?", tenantID).
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *GormUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Ensure GormUserRepository implements UserRepository
var _ UserRepository = (*GormUserRepository)(nil)
