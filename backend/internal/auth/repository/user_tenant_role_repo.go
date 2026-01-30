package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/gorm"
)

// UserTenantRoleRepository defines the interface for user-tenant-role data access.
type UserTenantRoleRepository interface {
	// FindByUserAndTenant retrieves the role assignment for a user in a tenant.
	FindByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.UserTenantRole, error)

	// Create creates a new user-tenant-role assignment.
	Create(ctx context.Context, role *domain.UserTenantRole) error

	// Update updates an existing user-tenant-role assignment.
	Update(ctx context.Context, role *domain.UserTenantRole) error

	// Delete removes a user-tenant-role assignment.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserAndTenant removes a user's role in a specific tenant.
	DeleteByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) error

	// ListByUser retrieves all role assignments for a user.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserTenantRole, error)

	// ListByTenant retrieves all role assignments for a tenant.
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.UserTenantRole, error)
}

// GormUserTenantRoleRepository is a GORM implementation of UserTenantRoleRepository.
type GormUserTenantRoleRepository struct {
	db *gorm.DB
}

// NewGormUserTenantRoleRepository creates a new GormUserTenantRoleRepository.
func NewGormUserTenantRoleRepository(db *gorm.DB) *GormUserTenantRoleRepository {
	return &GormUserTenantRoleRepository{db: db}
}

// FindByUserAndTenant retrieves the role assignment for a user in a tenant.
func (r *GormUserTenantRoleRepository) FindByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.UserTenantRole, error) {
	var role domain.UserTenantRole
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		First(&role, "user_id = ? AND tenant_id = ?", userID, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotInTenant
		}
		return nil, err
	}
	return &role, nil
}

// Create creates a new user-tenant-role assignment.
func (r *GormUserTenantRoleRepository) Create(ctx context.Context, role *domain.UserTenantRole) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(role).Error
}

// Update updates an existing user-tenant-role assignment.
func (r *GormUserTenantRoleRepository) Update(ctx context.Context, role *domain.UserTenantRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete removes a user-tenant-role assignment.
func (r *GormUserTenantRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.UserTenantRole{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotInTenant
	}
	return nil
}

// DeleteByUserAndTenant removes a user's role in a specific tenant.
func (r *GormUserTenantRoleRepository) DeleteByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.UserTenantRole{}, "user_id = ? AND tenant_id = ?", userID, tenantID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotInTenant
	}
	return nil
}

// ListByUser retrieves all role assignments for a user.
func (r *GormUserTenantRoleRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserTenantRole, error) {
	var roles []*domain.UserTenantRole
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// ListByTenant retrieves all role assignments for a tenant.
func (r *GormUserTenantRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.UserTenantRole, error) {
	var roles []*domain.UserTenantRole
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Ensure GormUserTenantRoleRepository implements UserTenantRoleRepository
var _ UserTenantRoleRepository = (*GormUserTenantRoleRepository)(nil)
