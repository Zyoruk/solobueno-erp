package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an active authentication session.
// Refresh tokens are stored here for revocation support.
type Session struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	RefreshToken string     `gorm:"uniqueIndex;size:255;not null" json:"-"` // Hashed token
	DeviceInfo   string     `gorm:"size:500" json:"device_info,omitempty"`
	IPAddress    string     `gorm:"size:45" json:"ip_address,omitempty"` // IPv6 max length
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt    time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt    *time.Time `gorm:"index" json:"revoked_at,omitempty"`

	// Associations
	User   User   `gorm:"foreignKey:UserID" json:"-"`
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

// TableName specifies the table name for GORM.
func (Session) TableName() string {
	return "sessions"
}

// IsValid checks if the session is still valid (not revoked and not expired).
func (s *Session) IsValid() bool {
	return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsRevoked checks if the session has been revoked.
func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

// Revoke marks the session as revoked.
func (s *Session) Revoke() {
	now := time.Now()
	s.RevokedAt = &now
}

// TimeUntilExpiry returns the duration until the session expires.
func (s *Session) TimeUntilExpiry() time.Duration {
	return time.Until(s.ExpiresAt)
}
