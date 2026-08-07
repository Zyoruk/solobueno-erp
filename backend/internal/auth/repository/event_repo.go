package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/gorm"
)

// AuthEventRepository defines the interface for auth event data access.
type AuthEventRepository interface {
	// Create creates a new auth event.
	Create(ctx context.Context, event *domain.AuthEvent) error

	// FindByUser retrieves auth events for a specific user.
	FindByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.AuthEvent, int64, error)

	// FindByTenant retrieves auth events for a specific tenant.
	FindByTenant(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*domain.AuthEvent, int64, error)

	// FindByType retrieves auth events of a specific type.
	FindByType(ctx context.Context, eventType domain.AuthEventType, offset, limit int) ([]*domain.AuthEvent, int64, error)

	// FindByUserAndType retrieves auth events for a user of a specific type.
	FindByUserAndType(ctx context.Context, userID uuid.UUID, eventType domain.AuthEventType, since time.Time) ([]*domain.AuthEvent, error)

	// CountRecentByIP counts recent events from a specific IP address.
	CountRecentByIP(ctx context.Context, ipAddress string, eventType domain.AuthEventType, since time.Time) (int64, error)

	// DeleteOlderThan removes events older than the specified time.
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// GormAuthEventRepository is a GORM implementation of AuthEventRepository.
type GormAuthEventRepository struct {
	db *gorm.DB
}

// NewGormAuthEventRepository creates a new GormAuthEventRepository.
func NewGormAuthEventRepository(db *gorm.DB) *GormAuthEventRepository {
	return &GormAuthEventRepository{db: db}
}

// Create creates a new auth event.
func (r *GormAuthEventRepository) Create(ctx context.Context, event *domain.AuthEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(event).Error
}

// FindByUser retrieves auth events for a specific user with pagination.
func (r *GormAuthEventRepository) FindByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.AuthEvent, int64, error) {
	var events []*domain.AuthEvent
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.AuthEvent{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// FindByTenant retrieves auth events for a specific tenant with pagination.
func (r *GormAuthEventRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*domain.AuthEvent, int64, error) {
	var events []*domain.AuthEvent
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.AuthEvent{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// FindByType retrieves auth events of a specific type with pagination.
func (r *GormAuthEventRepository) FindByType(ctx context.Context, eventType domain.AuthEventType, offset, limit int) ([]*domain.AuthEvent, int64, error) {
	var events []*domain.AuthEvent
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.AuthEvent{}).Where("event_type = ?", eventType).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Where("event_type = ?", eventType).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// FindByUserAndType retrieves auth events for a user of a specific type since a given time.
func (r *GormAuthEventRepository) FindByUserAndType(ctx context.Context, userID uuid.UUID, eventType domain.AuthEventType, since time.Time) ([]*domain.AuthEvent, error) {
	var events []*domain.AuthEvent
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND event_type = ? AND created_at >= ?", userID, eventType, since).
		Order("created_at DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// CountRecentByIP counts recent events from a specific IP address.
func (r *GormAuthEventRepository) CountRecentByIP(ctx context.Context, ipAddress string, eventType domain.AuthEventType, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.AuthEvent{}).
		Where("ip_address = ? AND event_type = ? AND created_at >= ?", ipAddress, eventType, since).
		Count(&count).Error
	return count, err
}

// DeleteOlderThan removes events older than the specified time.
func (r *GormAuthEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&domain.AuthEvent{})
	return result.RowsAffected, result.Error
}

// Ensure GormAuthEventRepository implements AuthEventRepository
var _ AuthEventRepository = (*GormAuthEventRepository)(nil)
