package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"gorm.io/gorm"
)

// PasswordResetRepository defines the interface for password reset token data access.
type PasswordResetRepository interface {
	// Create creates a new password reset token.
	Create(ctx context.Context, token *domain.PasswordResetToken) error

	// FindByToken retrieves a password reset token by its hash.
	FindByToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error)

	// MarkUsed marks a password reset token as used.
	MarkUsed(ctx context.Context, id uuid.UUID) error

	// DeleteExpired removes all expired tokens.
	DeleteExpired(ctx context.Context) (int64, error)

	// DeleteForUser removes all tokens for a specific user.
	DeleteForUser(ctx context.Context, userID uuid.UUID) error

	// CountRecentForUser counts recent (non-expired, non-used) tokens for a user.
	CountRecentForUser(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error)
}

// GormPasswordResetRepository is a GORM implementation of PasswordResetRepository.
type GormPasswordResetRepository struct {
	db *gorm.DB
}

// NewGormPasswordResetRepository creates a new GormPasswordResetRepository.
func NewGormPasswordResetRepository(db *gorm.DB) *GormPasswordResetRepository {
	return &GormPasswordResetRepository{db: db}
}

// Create creates a new password reset token.
func (r *GormPasswordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(token).Error
}

// FindByToken retrieves a password reset token by its hash.
func (r *GormPasswordResetRepository) FindByToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	var token domain.PasswordResetToken
	if err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPasswordResetInvalid
		}
		return nil, err
	}
	return &token, nil
}

// MarkUsed marks a password reset token as used.
func (r *GormPasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&domain.PasswordResetToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", time.Now())

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPasswordResetUsed
	}
	return nil
}

// DeleteExpired removes all expired tokens.
func (r *GormPasswordResetRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&domain.PasswordResetToken{})
	return result.RowsAffected, result.Error
}

// DeleteForUser removes all tokens for a specific user.
func (r *GormPasswordResetRepository) DeleteForUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.PasswordResetToken{}).Error
}

// CountRecentForUser counts recent tokens for a user.
func (r *GormPasswordResetRepository) CountRecentForUser(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.PasswordResetToken{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

// Ensure GormPasswordResetRepository implements PasswordResetRepository
var _ PasswordResetRepository = (*GormPasswordResetRepository)(nil)
