package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
)

func TestBaseEvent_OccurredAt(t *testing.T) {
	before := time.Now()
	event := newBaseEvent()
	after := time.Now()

	if event.occurredAt.Before(before) || event.occurredAt.After(after) {
		t.Error("occurredAt should be between before and after")
	}
}

func TestUserCreatedEvent(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()

	event := NewUserCreatedEvent(userID, "test@example.com", tenantID, domain.RoleManager, createdBy)

	if event.EventName() != "auth.user.created" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.user.created")
	}
	if event.UserID != userID {
		t.Error("UserID mismatch")
	}
	if event.Email != "test@example.com" {
		t.Error("Email mismatch")
	}
	if event.TenantID != tenantID {
		t.Error("TenantID mismatch")
	}
	if event.Role != domain.RoleManager {
		t.Error("Role mismatch")
	}
	if event.CreatedBy != createdBy {
		t.Error("CreatedBy mismatch")
	}
	if event.OccurredAt().IsZero() {
		t.Error("OccurredAt should be set")
	}
}

func TestLoginSucceededEvent(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	event := NewLoginSucceededEvent(userID, tenantID, "192.168.1.1", "Mozilla/5.0")

	if event.EventName() != "auth.login.succeeded" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.login.succeeded")
	}
	if event.UserID != userID {
		t.Error("UserID mismatch")
	}
	if event.TenantID != tenantID {
		t.Error("TenantID mismatch")
	}
	if event.IPAddress != "192.168.1.1" {
		t.Error("IPAddress mismatch")
	}
	if event.UserAgent != "Mozilla/5.0" {
		t.Error("UserAgent mismatch")
	}
}

func TestLoginFailedEvent(t *testing.T) {
	event := NewLoginFailedEvent("test@example.com", "192.168.1.1", "Mozilla/5.0", "invalid_password")

	if event.EventName() != "auth.login.failed" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.login.failed")
	}
	if event.Email != "test@example.com" {
		t.Error("Email mismatch")
	}
	if event.Reason != "invalid_password" {
		t.Error("Reason mismatch")
	}
}

func TestLogoutEvent(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	event := NewLogoutEvent(userID, sessionID, "192.168.1.1")

	if event.EventName() != "auth.logout" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.logout")
	}
	if event.UserID != userID {
		t.Error("UserID mismatch")
	}
	if event.SessionID != sessionID {
		t.Error("SessionID mismatch")
	}
}

func TestTokenRefreshedEvent(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	oldSessionID := uuid.New()
	newSessionID := uuid.New()

	event := NewTokenRefreshedEvent(userID, tenantID, oldSessionID, newSessionID)

	if event.EventName() != "auth.token.refreshed" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.token.refreshed")
	}
	if event.OldSessionID != oldSessionID {
		t.Error("OldSessionID mismatch")
	}
	if event.NewSessionID != newSessionID {
		t.Error("NewSessionID mismatch")
	}
}

func TestPasswordChangedEvent(t *testing.T) {
	userID := uuid.New()

	event := NewPasswordChangedEvent(userID, "192.168.1.1")

	if event.EventName() != "auth.password.changed" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.password.changed")
	}
	if event.UserID != userID {
		t.Error("UserID mismatch")
	}
	if event.IPAddress != "192.168.1.1" {
		t.Error("IPAddress mismatch")
	}
}

func TestRoleChangedEvent(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	changedBy := uuid.New()

	event := NewRoleChangedEvent(userID, tenantID, domain.RoleWaiter, domain.RoleManager, changedBy)

	if event.EventName() != "auth.role.changed" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.role.changed")
	}
	if event.OldRole != domain.RoleWaiter {
		t.Error("OldRole mismatch")
	}
	if event.NewRole != domain.RoleManager {
		t.Error("NewRole mismatch")
	}
	if event.ChangedBy != changedBy {
		t.Error("ChangedBy mismatch")
	}
}

func TestSessionRevokedEvent(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	revokedBy := uuid.New()

	event := NewSessionRevokedEvent(userID, sessionID, revokedBy, "password_changed")

	if event.EventName() != "auth.session.revoked" {
		t.Errorf("EventName() = %q, want %q", event.EventName(), "auth.session.revoked")
	}
	if event.UserID != userID {
		t.Error("UserID mismatch")
	}
	if event.SessionID != sessionID {
		t.Error("SessionID mismatch")
	}
	if event.RevokedBy != revokedBy {
		t.Error("RevokedBy mismatch")
	}
	if event.Reason != "password_changed" {
		t.Error("Reason mismatch")
	}
}
