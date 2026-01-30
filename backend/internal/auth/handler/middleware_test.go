package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantErr   bool
	}{
		{
			name:      "valid bearer token",
			header:    "Bearer abc123",
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name:      "lowercase bearer",
			header:    "bearer abc123",
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name:      "missing header",
			header:    "",
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "wrong scheme",
			header:    "Basic abc123",
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "no token after bearer",
			header:    "Bearer",
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			token, err := extractBearerToken(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if token != tt.wantToken {
				t.Errorf("extractBearerToken() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		xRealIP    string
		remoteAddr string
		want       string
	}{
		{
			name:       "X-Forwarded-For single",
			xff:        "192.168.1.1",
			remoteAddr: "10.0.0.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For multiple",
			xff:        "192.168.1.1, 10.0.0.2, 172.16.0.1",
			remoteAddr: "10.0.0.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "X-Real-IP",
			xRealIP:    "192.168.1.2",
			remoteAddr: "10.0.0.1:12345",
			want:       "192.168.1.2",
		},
		{
			name:       "RemoteAddr only",
			remoteAddr: "10.0.0.1:12345",
			want:       "10.0.0.1",
		},
		{
			name:       "RemoteAddr no port",
			remoteAddr: "10.0.0.1",
			want:       "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			req.RemoteAddr = tt.remoteAddr

			if got := GetClientIP(req); got != tt.want {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextHelpers(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	role := domain.RoleManager

	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDContextKey, userID)
	ctx = context.WithValue(ctx, TenantIDContextKey, tenantID)
	ctx = context.WithValue(ctx, RoleContextKey, role)

	// Test GetUserID
	gotUserID, ok := GetUserID(ctx)
	if !ok {
		t.Error("GetUserID() should return true")
	}
	if gotUserID != userID {
		t.Errorf("GetUserID() = %v, want %v", gotUserID, userID)
	}

	// Test GetTenantID
	gotTenantID, ok := GetTenantID(ctx)
	if !ok {
		t.Error("GetTenantID() should return true")
	}
	if gotTenantID != tenantID {
		t.Errorf("GetTenantID() = %v, want %v", gotTenantID, tenantID)
	}

	// Test GetRole
	gotRole, ok := GetRole(ctx)
	if !ok {
		t.Error("GetRole() should return true")
	}
	if gotRole != role {
		t.Errorf("GetRole() = %v, want %v", gotRole, role)
	}

	// Test with empty context
	emptyCtx := context.Background()
	_, ok = GetUserID(emptyCtx)
	if ok {
		t.Error("GetUserID() on empty context should return false")
	}
}

func TestGetClaims(t *testing.T) {
	claims := &domain.Claims{
		TenantID: uuid.New(),
		Role:     domain.RoleAdmin,
		Email:    "test@example.com",
	}

	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	gotClaims, ok := GetClaims(ctx)
	if !ok {
		t.Error("GetClaims() should return true")
	}
	if gotClaims.Email != claims.Email {
		t.Errorf("GetClaims().Email = %q, want %q", gotClaims.Email, claims.Email)
	}

	// Test with empty context
	_, ok = GetClaims(context.Background())
	if ok {
		t.Error("GetClaims() on empty context should return false")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	data := map[string]string{"message": "hello"}
	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	expected := `{"message":"hello"}`
	// Note: json.Encoder adds a newline
	if w.Body.String() != expected+"\n" {
		t.Errorf("Body = %q, want %q", w.Body.String(), expected)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "invalid_request", "Something went wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	// Check that body contains error code
	body := w.Body.String()
	if body == "" {
		t.Error("Body should not be empty")
	}
}
