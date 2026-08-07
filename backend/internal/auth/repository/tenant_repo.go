package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/gorm"
)

// TenantRepository defines the interface for tenant data access.
type TenantRepository interface {
	// FindByID retrieves a tenant by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)

	// FindBySlug retrieves a tenant by its slug.
	FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error)

	// Create creates a new tenant.
	Create(ctx context.Context, tenant *domain.Tenant) error

	// Update updates an existing tenant.
	Update(ctx context.Context, tenant *domain.Tenant) error

	// ExistsBySlug checks if a tenant with the given slug exists.
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
}

// GormTenantRepository is a GORM implementation of TenantRepository.
type GormTenantRepository struct {
	db *gorm.DB
}

// NewGormTenantRepository creates a new GormTenantRepository.
func NewGormTenantRepository(db *gorm.DB) *GormTenantRepository {
	return &GormTenantRepository{db: db}
}

// FindByID retrieves a tenant by its ID.
func (r *GormTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// FindBySlug retrieves a tenant by its slug.
func (r *GormTenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// Create creates a new tenant.
func (r *GormTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(tenant).Error
}

// Update updates an existing tenant.
func (r *GormTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

// ExistsBySlug checks if a tenant with the given slug exists.
func (r *GormTenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Tenant{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Ensure GormTenantRepository implements TenantRepository
var _ TenantRepository = (*GormTenantRepository)(nil)
