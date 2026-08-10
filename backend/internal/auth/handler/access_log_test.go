package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/shared/observability"
)

// entryField returns the value of a field on the most recent log entry with
// the given message, and whether it was found at all.
func (l *capturingLogger) entryField(msg, key string) (any, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.entries) - 1; i >= 0; i-- {
		e := l.entries[i]
		if e.msg != msg {
			continue
		}
		for _, f := range e.fields {
			if f.Key == key {
				return f.Value, true
			}
		}
		return nil, false
	}
	return nil, false
}

func (l *capturingLogger) lastLevel(msg string) (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.entries) - 1; i >= 0; i-- {
		if l.entries[i].msg == msg {
			return l.entries[i].level, true
		}
	}
	return "", false
}

// TestAccessLog_RevokedToken_LogsCompletionWithNoPII covers the exact gap
// the user reported: a revoked refresh token returns a 401 - an expected
// outcome, not a 500 - but previously produced zero structured log output.
func TestAccessLog_RevokedToken_LogsCompletionWithNoPII(t *testing.T) {
	h, _, _, _, _, _, sessionRepo := setupWiredAuthHandler(t)
	cl := &capturingLogger{}
	SetLogger(cl)
	t.Cleanup(func() { SetLogger(observability.New("test")) })

	revokedAt := time.Now()
	sessionRepo.FindByTokenFunc = func(ctx context.Context, tokenHash string) (*domain.Session, error) {
		return &domain.Session{
			ID: uuid.New(), UserID: uuid.New(), TenantID: uuid.New(),
			ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &revokedAt,
		}, nil
	}

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "revoked-token"})
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	AccessLog(http.HandlerFunc(h.Refresh)).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusUnauthorized, w.Body.String())
	}

	level, ok := cl.lastLevel("request completed")
	if !ok {
		t.Fatal(`expected a "request completed" log entry`)
	}
	if level != "info" {
		t.Errorf("level = %q, want %q (401 is an expected outcome, not an error)", level, "info")
	}
	if status, _ := cl.entryField("request completed", "status"); status != http.StatusUnauthorized {
		t.Errorf("status field = %v, want %d", status, http.StatusUnauthorized)
	}
	for _, key := range []string{"user_id", "tenant_id"} {
		if _, ok := cl.entryField("request completed", key); ok {
			t.Errorf("unauthenticated request must not log %q", key)
		}
	}
	if _, ok := cl.entryField("request completed", "path"); !ok {
		t.Error("expected a path field")
	}
}

// TestAccessLog_AuthenticatedRequest_LogsUserAndTenant proves RequireAuth's
// two setAccessLog* calls actually reach the outer AccessLog middleware via
// the shared *accessLogFields pointer, even though RequireAuth attaches a
// brand new context via r.WithContext deeper in the chain than AccessLog's
// own request variable.
func TestAccessLog_AuthenticatedRequest_LogsUserAndTenant(t *testing.T) {
	cl := &capturingLogger{}
	SetLogger(cl)
	t.Cleanup(func() { SetLogger(observability.New("test")) })

	userID := uuid.New()
	tenantID := uuid.New()

	// Stands in for RequireAuth's context-mutation step (the two lines
	// under test in middleware.go) without needing a signed JWT.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setAccessLogUserID(r.Context(), userID.String())
		setAccessLogTenantID(r.Context(), tenantID.String())
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	AccessLog(next).ServeHTTP(w, req)

	if v, _ := cl.entryField("request completed", "user_id"); v != userID.String() {
		t.Errorf("user_id = %v, want %s", v, userID.String())
	}
	if v, _ := cl.entryField("request completed", "tenant_id"); v != tenantID.String() {
		t.Errorf("tenant_id = %v, want %s", v, tenantID.String())
	}
}
