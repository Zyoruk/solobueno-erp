package domain

import (
	"testing"
	"time"
)

func TestPasswordResetToken_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		token PasswordResetToken
		want  bool
	}{
		{
			name: "valid token",
			token: PasswordResetToken{
				ExpiresAt: now.Add(time.Hour),
				UsedAt:    nil,
			},
			want: true,
		},
		{
			name: "expired token",
			token: PasswordResetToken{
				ExpiresAt: now.Add(-time.Hour),
				UsedAt:    nil,
			},
			want: false,
		},
		{
			name: "used token",
			token: PasswordResetToken{
				ExpiresAt: now.Add(time.Hour),
				UsedAt:    &now,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	now := time.Now()

	t1 := PasswordResetToken{ExpiresAt: now.Add(time.Hour)}
	if t1.IsExpired() {
		t.Error("Future expiry should not be expired")
	}

	t2 := PasswordResetToken{ExpiresAt: now.Add(-time.Hour)}
	if !t2.IsExpired() {
		t.Error("Past expiry should be expired")
	}
}

func TestPasswordResetToken_IsUsed(t *testing.T) {
	now := time.Now()

	t1 := PasswordResetToken{UsedAt: nil}
	if t1.IsUsed() {
		t.Error("nil UsedAt should not be used")
	}

	t2 := PasswordResetToken{UsedAt: &now}
	if !t2.IsUsed() {
		t.Error("non-nil UsedAt should be used")
	}
}

func TestPasswordResetToken_MarkUsed(t *testing.T) {
	token := PasswordResetToken{UsedAt: nil}

	if token.IsUsed() {
		t.Error("Should not be used initially")
	}

	token.MarkUsed()

	if !token.IsUsed() {
		t.Error("Should be used after MarkUsed()")
	}
}

func TestPasswordResetToken_TableName(t *testing.T) {
	token := PasswordResetToken{}
	if token.TableName() != "password_reset_tokens" {
		t.Errorf("TableName() = %q, want %q", token.TableName(), "password_reset_tokens")
	}
}
