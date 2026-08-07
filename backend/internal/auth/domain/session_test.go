package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSession_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		session Session
		want    bool
	}{
		{
			name: "valid session",
			session: Session{
				ExpiresAt: now.Add(time.Hour),
				RevokedAt: nil,
			},
			want: true,
		},
		{
			name: "expired session",
			session: Session{
				ExpiresAt: now.Add(-time.Hour),
				RevokedAt: nil,
			},
			want: false,
		},
		{
			name: "revoked session",
			session: Session{
				ExpiresAt: now.Add(time.Hour),
				RevokedAt: &now,
			},
			want: false,
		},
		{
			name: "expired and revoked",
			session: Session{
				ExpiresAt: now.Add(-time.Hour),
				RevokedAt: &now,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.session.IsValid(); got != tt.want {
				t.Errorf("Session.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_IsExpired(t *testing.T) {
	now := time.Now()

	s1 := Session{ExpiresAt: now.Add(time.Hour)}
	if s1.IsExpired() {
		t.Error("Future expiry should not be expired")
	}

	s2 := Session{ExpiresAt: now.Add(-time.Hour)}
	if !s2.IsExpired() {
		t.Error("Past expiry should be expired")
	}
}

func TestSession_IsRevoked(t *testing.T) {
	now := time.Now()

	s1 := Session{RevokedAt: nil}
	if s1.IsRevoked() {
		t.Error("nil RevokedAt should not be revoked")
	}

	s2 := Session{RevokedAt: &now}
	if !s2.IsRevoked() {
		t.Error("non-nil RevokedAt should be revoked")
	}
}

func TestSession_Revoke(t *testing.T) {
	s := Session{
		ID:        uuid.New(),
		RevokedAt: nil,
	}

	if s.IsRevoked() {
		t.Error("Should not be revoked initially")
	}

	s.Revoke()

	if !s.IsRevoked() {
		t.Error("Should be revoked after Revoke()")
	}

	if s.RevokedAt == nil {
		t.Error("RevokedAt should be set")
	}
}

func TestSession_TimeUntilExpiry(t *testing.T) {
	s := Session{
		ExpiresAt: time.Now().Add(time.Hour),
	}

	duration := s.TimeUntilExpiry()
	if duration < 59*time.Minute || duration > 61*time.Minute {
		t.Errorf("TimeUntilExpiry() = %v, expected ~1 hour", duration)
	}
}

func TestSession_TableName(t *testing.T) {
	s := Session{}
	if s.TableName() != "sessions" {
		t.Errorf("TableName() = %q, want %q", s.TableName(), "sessions")
	}
}
