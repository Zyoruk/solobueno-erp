package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuthEventType represents the type of authentication event.
type AuthEventType string

// AuthEventType constants define all authentication event types.
const (
	EventLoginSuccess           AuthEventType = "login_success"
	EventLoginFailed            AuthEventType = "login_failed"
	EventLogout                 AuthEventType = "logout"
	EventTokenRefresh           AuthEventType = "token_refresh"
	EventPasswordChanged        AuthEventType = "password_changed"
	EventPasswordResetRequested AuthEventType = "password_reset_requested"
	EventPasswordResetCompleted AuthEventType = "password_reset_completed"
	EventAccountCreated         AuthEventType = "account_created"
	EventAccountDisabled        AuthEventType = "account_disabled"
	EventAccountEnabled         AuthEventType = "account_enabled"
	EventRoleChanged            AuthEventType = "role_changed"
	EventSessionRevoked         AuthEventType = "session_revoked"
)

// String returns the string representation of the event type.
func (t AuthEventType) String() string {
	return string(t)
}

// GormDataType implements GORM's custom type interface.
func (t AuthEventType) GormDataType() string {
	return "varchar(50)"
}

// Metadata represents flexible JSON metadata for auth events.
type Metadata map[string]interface{}

// Scan implements sql.Scanner interface for database reads.
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan type %T into Metadata", value)
	}

	if len(bytes) == 0 {
		*m = nil
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer interface for database writes.
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// GormDataType implements GORM's custom type interface.
func (m Metadata) GormDataType() string {
	return "jsonb"
}

// AuthEvent represents an audit log entry for authentication actions.
type AuthEvent struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    *uuid.UUID    `gorm:"type:uuid;index" json:"user_id,omitempty"`
	TenantID  *uuid.UUID    `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	EventType AuthEventType `gorm:"size:50;not null;index" json:"event_type"`
	IPAddress string        `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent string        `gorm:"size:500" json:"user_agent,omitempty"`
	Metadata  Metadata      `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time     `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName specifies the table name for GORM.
func (AuthEvent) TableName() string {
	return "auth_events"
}

// NewAuthEvent creates a new auth event with the given parameters.
func NewAuthEvent(eventType AuthEventType, userID, tenantID *uuid.UUID, ipAddress, userAgent string) *AuthEvent {
	return &AuthEvent{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		EventType: eventType,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}
}

// WithMetadata adds metadata to the event and returns the event for chaining.
func (e *AuthEvent) WithMetadata(key string, value interface{}) *AuthEvent {
	if e.Metadata == nil {
		e.Metadata = make(Metadata)
	}
	e.Metadata[key] = value
	return e
}
