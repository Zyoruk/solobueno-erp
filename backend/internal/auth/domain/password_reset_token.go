package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a time-limited token for password reset flow.
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"uniqueIndex;size:255;not null" json:"-"` // Hashed token
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`

	// Associations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for GORM.
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsValid checks if the token is still valid (not used and not expired).
func (t *PasswordResetToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}

// IsExpired checks if the token has expired.
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used.
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// MarkUsed marks the token as used.
func (t *PasswordResetToken) MarkUsed() {
	now := time.Now()
	t.UsedAt = &now
}
