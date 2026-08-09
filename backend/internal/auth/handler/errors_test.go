package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/shared/observability"
)

// capturingLogger is a test double for observability.Logger that records
// every call, so tests can assert an error was actually logged - proving
// writeInternalError's log line survives, not just its generic HTTP response.
type capturingLogger struct {
	mu      sync.Mutex
	entries []capturedEntry
}

type capturedEntry struct {
	level  string
	msg    string
	fields []observability.Field
}

func (l *capturingLogger) Debug(msg string, fields ...observability.Field) {
	l.add("debug", msg, fields)
}
func (l *capturingLogger) Info(msg string, fields ...observability.Field) { l.add("info", msg, fields) }
func (l *capturingLogger) Warn(msg string, fields ...observability.Field) { l.add("warn", msg, fields) }
func (l *capturingLogger) Error(msg string, fields ...observability.Field) {
	l.add("error", msg, fields)
}
func (l *capturingLogger) With(fields ...observability.Field) observability.Logger {
	return l
}

func (l *capturingLogger) add(level, msg string, fields []observability.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, capturedEntry{level: level, msg: msg, fields: fields})
}

// hasErrorContaining reports whether an Error-level entry has a field whose
// string value contains substr (e.g. the wrapped error message).
func (l *capturingLogger) hasErrorContaining(substr string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.entries {
		if e.level != "error" {
			continue
		}
		for _, f := range e.fields {
			if s, ok := f.Value.(string); ok && strings.Contains(s, substr) {
				return true
			}
		}
	}
	return false
}

var _ observability.Logger = (*capturingLogger)(nil)

func TestAuthHandler_Refresh_LogsSessionLookupFailure(t *testing.T) {
	h, _, _, _, _, _, sessionRepo := setupWiredAuthHandler(t)
	cl := &capturingLogger{}
	SetLogger(cl)
	t.Cleanup(func() { SetLogger(observability.New("test")) })

	sessionRepo.FindByTokenFunc = func(ctx context.Context, tokenHash string) (*domain.Session, error) {
		return nil, errors.New("db down")
	}

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "sometoken"})
	req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "db down") {
		t.Errorf("response must never leak internal error detail: %s", w.Body.String())
	}
	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "internal_error" {
		t.Errorf("Code = %q, want %q", resp.Error.Code, "internal_error")
	}
	if !cl.hasErrorContaining("session lookup") {
		t.Error("expected an error-level log entry mentioning the session lookup failure")
	}
}

func TestAuthHandler_Refresh_LogsUserLookupFailure(t *testing.T) {
	h, _, _, userRepo, _, _, sessionRepo := setupWiredAuthHandler(t)
	cl := &capturingLogger{}
	SetLogger(cl)
	t.Cleanup(func() { SetLogger(observability.New("test")) })

	userID := uuid.New()
	sessionRepo.FindByTokenFunc = func(ctx context.Context, tokenHash string) (*domain.Session, error) {
		return &domain.Session{ID: uuid.New(), UserID: userID, TenantID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	userRepo.FindByIDFunc = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
		return nil, errors.New("db down")
	}

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "sometoken"})
	req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "db down") {
		t.Errorf("response must never leak internal error detail: %s", w.Body.String())
	}
	if !cl.hasErrorContaining("user lookup") {
		t.Error("expected an error-level log entry mentioning the user lookup failure")
	}
}
