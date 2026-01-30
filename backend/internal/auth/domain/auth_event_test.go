package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestAuthEventType_String(t *testing.T) {
	if EventLoginSuccess.String() != "login_success" {
		t.Errorf("EventLoginSuccess.String() = %q, want %q", EventLoginSuccess.String(), "login_success")
	}
}

func TestAuthEventType_GormDataType(t *testing.T) {
	if EventLoginSuccess.GormDataType() != "varchar(50)" {
		t.Errorf("GormDataType() = %q, want %q", EventLoginSuccess.GormDataType(), "varchar(50)")
	}
}

func TestMetadata_ScanAndValue(t *testing.T) {
	var m Metadata

	// Test Scan with valid JSON
	if err := m.Scan([]byte(`{"key": "value", "num": 123}`)); err != nil {
		t.Errorf("Scan() error: %v", err)
	}
	if m["key"] != "value" {
		t.Errorf("m[key] = %v, want %q", m["key"], "value")
	}

	// Test Scan with nil
	if err := m.Scan(nil); err != nil {
		t.Errorf("Scan(nil) error: %v", err)
	}

	// Test Scan with empty bytes
	if err := m.Scan([]byte{}); err != nil {
		t.Errorf("Scan(empty) error: %v", err)
	}

	// Test Scan with invalid type
	if err := m.Scan(123); err == nil {
		t.Error("Scan(int) should return error")
	}

	// Test Value
	m = Metadata{"test": "data"}
	val, err := m.Value()
	if err != nil {
		t.Errorf("Value() error: %v", err)
	}
	if val == nil {
		t.Error("Value() should not be nil")
	}

	// Test Value with nil metadata
	var nilM Metadata
	val, err = nilM.Value()
	if err != nil {
		t.Errorf("nil Metadata Value() error: %v", err)
	}
	if val != nil {
		t.Error("nil Metadata Value() should be nil")
	}
}

func TestMetadata_GormDataType(t *testing.T) {
	m := Metadata{}
	if m.GormDataType() != "jsonb" {
		t.Errorf("GormDataType() = %q, want %q", m.GormDataType(), "jsonb")
	}
}

func TestNewAuthEvent(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	event := NewAuthEvent(EventLoginSuccess, &userID, &tenantID, "127.0.0.1", "Mozilla/5.0")

	if event.ID == uuid.Nil {
		t.Error("Event ID should be set")
	}
	if *event.UserID != userID {
		t.Error("UserID mismatch")
	}
	if *event.TenantID != tenantID {
		t.Error("TenantID mismatch")
	}
	if event.EventType != EventLoginSuccess {
		t.Error("EventType mismatch")
	}
	if event.IPAddress != "127.0.0.1" {
		t.Error("IPAddress mismatch")
	}
	if event.UserAgent != "Mozilla/5.0" {
		t.Error("UserAgent mismatch")
	}
	if event.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestAuthEvent_WithMetadata(t *testing.T) {
	event := NewAuthEvent(EventLoginFailed, nil, nil, "", "")

	result := event.WithMetadata("reason", "invalid_password")

	// Should return same event for chaining
	if result != event {
		t.Error("WithMetadata should return same event")
	}

	if event.Metadata["reason"] != "invalid_password" {
		t.Errorf("Metadata[reason] = %v, want %q", event.Metadata["reason"], "invalid_password")
	}

	// Add another
	event.WithMetadata("attempts", 3)
	if event.Metadata["attempts"] != 3 {
		t.Error("Should be able to add multiple metadata")
	}
}

func TestAuthEvent_TableName(t *testing.T) {
	e := AuthEvent{}
	if e.TableName() != "auth_events" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "auth_events")
	}
}
