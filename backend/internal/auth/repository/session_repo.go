package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/gorm"
)

// SessionRepository defines the interface for session data access.
type SessionRepository interface {
	// Create creates a new session.
	Create(ctx context.Context, session *domain.Session) error

	// FindByToken retrieves a session by its refresh token hash.
	FindByToken(ctx context.Context, tokenHash string) (*domain.Session, error)

	// FindByID retrieves a session by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)

	// Revoke revokes a session by setting its revoked_at timestamp.
	Revoke(ctx context.Context, id uuid.UUID) error

	// RevokeByToken revokes a session by its refresh token hash.
	RevokeByToken(ctx context.Context, tokenHash string) error

	// RevokeAllForUser revokes all sessions for a user.
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error

	// RevokeAllForUserInTenant revokes all sessions for a user in a specific tenant.
	RevokeAllForUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) error

	// DeleteExpired removes all expired sessions.
	DeleteExpired(ctx context.Context) (int64, error)

	// CountActiveForUser returns the number of active sessions for a user.
	CountActiveForUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

// GormSessionRepository is a GORM implementation of SessionRepository.
type GormSessionRepository struct {
	db *gorm.DB
}

// NewGormSessionRepository creates a new GormSessionRepository.
func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{db: db}
}

// Create creates a new session.
func (r *GormSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByToken retrieves a session by its refresh token hash.
func (r *GormSessionRepository) FindByToken(ctx context.Context, tokenHash string) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.WithContext(ctx).First(&session, "refresh_token = ?", tokenHash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

// FindByID retrieves a session by its ID.
func (r *GormSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

// Revoke revokes a session by setting its revoked_at timestamp.
func (r *GormSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", time.Now())

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

// RevokeByToken revokes a session by its refresh token hash.
func (r *GormSessionRepository) RevokeByToken(ctx context.Context, tokenHash string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("refresh_token = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", time.Now())

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

// RevokeAllForUser revokes all sessions for a user.
func (r *GormSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now()).Error
}

// RevokeAllForUserInTenant revokes all sessions for a user in a specific tenant.
func (r *GormSessionRepository) RevokeAllForUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("user_id = ? AND tenant_id = ? AND revoked_at IS NULL", userID, tenantID).
		Update("revoked_at", time.Now()).Error
}

// DeleteExpired removes all expired sessions.
func (r *GormSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&domain.Session{})
	return result.RowsAffected, result.Error
}

// CountActiveForUser returns the number of active sessions for a user.
func (r *GormSessionRepository) CountActiveForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Session{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}

// Ensure GormSessionRepository implements SessionRepository
var _ SessionRepository = (*GormSessionRepository)(nil)
