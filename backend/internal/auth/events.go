// Package auth provides authentication and authorization functionality.
package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
)

// DomainEvent represents a domain event that can be published.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// BaseEvent provides common fields for all domain events.
type BaseEvent struct {
	occurredAt time.Time
}

// OccurredAt returns when the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func newBaseEvent() BaseEvent {
	return BaseEvent{occurredAt: time.Now()}
}

// UserCreatedEvent is published when a new user is created.
type UserCreatedEvent struct {
	BaseEvent
	UserID    uuid.UUID
	Email     string
	TenantID  uuid.UUID
	Role      domain.Role
	CreatedBy uuid.UUID
}

// EventName returns the event name.
func (e UserCreatedEvent) EventName() string {
	return "auth.user.created"
}

// NewUserCreatedEvent creates a new UserCreatedEvent.
func NewUserCreatedEvent(userID uuid.UUID, email string, tenantID uuid.UUID, role domain.Role, createdBy uuid.UUID) UserCreatedEvent {
	return UserCreatedEvent{
		BaseEvent: newBaseEvent(),
		UserID:    userID,
		Email:     email,
		TenantID:  tenantID,
		Role:      role,
		CreatedBy: createdBy,
	}
}

// LoginSucceededEvent is published when a user successfully logs in.
type LoginSucceededEvent struct {
	BaseEvent
	UserID    uuid.UUID
	TenantID  uuid.UUID
	IPAddress string
	UserAgent string
}

// EventName returns the event name.
func (e LoginSucceededEvent) EventName() string {
	return "auth.login.succeeded"
}

// NewLoginSucceededEvent creates a new LoginSucceededEvent.
func NewLoginSucceededEvent(userID, tenantID uuid.UUID, ipAddress, userAgent string) LoginSucceededEvent {
	return LoginSucceededEvent{
		BaseEvent: newBaseEvent(),
		UserID:    userID,
		TenantID:  tenantID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// LoginFailedEvent is published when a login attempt fails.
type LoginFailedEvent struct {
	BaseEvent
	Email     string
	IPAddress string
	UserAgent string
	Reason    string
}

// EventName returns the event name.
func (e LoginFailedEvent) EventName() string {
	return "auth.login.failed"
}

// NewLoginFailedEvent creates a new LoginFailedEvent.
func NewLoginFailedEvent(email, ipAddress, userAgent, reason string) LoginFailedEvent {
	return LoginFailedEvent{
		BaseEvent: newBaseEvent(),
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Reason:    reason,
	}
}

// LogoutEvent is published when a user logs out.
type LogoutEvent struct {
	BaseEvent
	UserID    uuid.UUID
	SessionID uuid.UUID
	IPAddress string
}

// EventName returns the event name.
func (e LogoutEvent) EventName() string {
	return "auth.logout"
}

// NewLogoutEvent creates a new LogoutEvent.
func NewLogoutEvent(userID, sessionID uuid.UUID, ipAddress string) LogoutEvent {
	return LogoutEvent{
		BaseEvent: newBaseEvent(),
		UserID:    userID,
		SessionID: sessionID,
		IPAddress: ipAddress,
	}
}

// TokenRefreshedEvent is published when a token is refreshed.
type TokenRefreshedEvent struct {
	BaseEvent
	UserID       uuid.UUID
	TenantID     uuid.UUID
	OldSessionID uuid.UUID
	NewSessionID uuid.UUID
}

// EventName returns the event name.
func (e TokenRefreshedEvent) EventName() string {
	return "auth.token.refreshed"
}

// NewTokenRefreshedEvent creates a new TokenRefreshedEvent.
func NewTokenRefreshedEvent(userID, tenantID, oldSessionID, newSessionID uuid.UUID) TokenRefreshedEvent {
	return TokenRefreshedEvent{
		BaseEvent:    newBaseEvent(),
		UserID:       userID,
		TenantID:     tenantID,
		OldSessionID: oldSessionID,
		NewSessionID: newSessionID,
	}
}

// PasswordChangedEvent is published when a user changes their password.
type PasswordChangedEvent struct {
	BaseEvent
	UserID    uuid.UUID
	IPAddress string
}

// EventName returns the event name.
func (e PasswordChangedEvent) EventName() string {
	return "auth.password.changed"
}

// NewPasswordChangedEvent creates a new PasswordChangedEvent.
func NewPasswordChangedEvent(userID uuid.UUID, ipAddress string) PasswordChangedEvent {
	return PasswordChangedEvent{
		BaseEvent: newBaseEvent(),
		UserID:    userID,
		IPAddress: ipAddress,
	}
}

// RoleChangedEvent is published when a user's role is changed.
type RoleChangedEvent struct {
	BaseEvent
	UserID    uuid.UUID
	TenantID  uuid.UUID
	OldRole   domain.Role
	NewRole   domain.Role
	ChangedBy uuid.UUID
}

// EventName returns the event name.
func (e RoleChangedEvent) EventName() string {
	return "auth.role.changed"
}

// NewRoleChangedEvent creates a new RoleChangedEvent.
func NewRoleChangedEvent(userID, tenantID uuid.UUID, oldRole, newRole domain.Role, changedBy uuid.UUID) RoleChangedEvent {
	return RoleChangedEvent{
		BaseEvent: newBaseEvent(),
		UserID:    userID,
		TenantID:  tenantID,
		OldRole:   oldRole,
		NewRole:   newRole,
		ChangedBy: changedBy,
	}
}

// SessionRevokedEvent is published when a session is revoked.
type SessionRevokedEvent struct {
	BaseEvent
	UserID    uuid.UUID
	SessionID uuid.UUID
	RevokedBy uuid.UUID
	Reason    string
}

// EventName returns the event name.
func (e SessionRevokedEvent) EventName() string {
	return "auth.session.revoked"
}

// NewSessionRevokedEvent creates a new SessionRevokedEvent.
func NewSessionRevokedEvent(userID, sessionID, revokedBy uuid.UUID, reason string) SessionRevokedEvent {
	return SessionRevokedEvent{
		BaseEvent: newBaseEvent(),
		UserID:    userID,
		SessionID: sessionID,
		RevokedBy: revokedBy,
		Reason:    reason,
	}
}
