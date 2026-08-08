package service

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
)

// p95 returns the 95th-percentile duration from a set of samples.
func p95(samples []time.Duration) time.Duration {
	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(float64(len(sorted))*0.95) - 1
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}

// TestLoginPerformance guards SC-001: login must complete within 500ms
// under normal load. Argon2id verification is the dominant cost, so this
// catches regressions in hashing parameters or accidental N+1 lookups.
func TestLoginPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	authSvc, userRepo, _, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	passwordHash, err := NewPasswordService().Hash("Password123!")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	userRepo.AddUser(&domain.User{
		ID: userID, Email: "perf-login@example.com", PasswordHash: passwordHash, IsActive: true,
		TenantRoles: []domain.UserTenantRole{{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleManager}},
	})
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, Name: "Acme", Slug: "acme", IsActive: true})

	const iterations = 20
	samples := make([]time.Duration, iterations)
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := authSvc.Login(ctx, LoginRequest{Email: "perf-login@example.com", Password: "Password123!"})
		samples[i] = time.Since(start)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
	}

	got := p95(samples)
	want := 500 * time.Millisecond
	if got > want {
		t.Errorf("Login p95 latency = %v, want <= %v (SC-001)", got, want)
	}
}

// TestRefreshPerformance guards SC-002: token refresh must complete within
// 200ms — no password hashing on this path, just token generation and a
// session rotation.
func TestRefreshPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	authSvc, userRepo, sessionRepo, tenantRepo, _ := setupAuthService(t)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	userRepo.AddUser(&domain.User{
		ID: userID, Email: "perf-refresh@example.com", IsActive: true,
		TenantRoles: []domain.UserTenantRole{{ID: uuid.New(), UserID: userID, TenantID: tenantID, Role: domain.RoleManager}},
	})
	tenantRepo.AddTenant(&domain.Tenant{ID: tenantID, Name: "Acme", Slug: "acme", IsActive: true})

	const iterations = 20
	samples := make([]time.Duration, iterations)
	for i := 0; i < iterations; i++ {
		plainToken := uuid.New().String()
		sessionRepo.Create(ctx, &domain.Session{
			ID: uuid.New(), UserID: userID, TenantID: tenantID,
			RefreshToken: authSvc.tokenService.HashRefreshToken(plainToken),
			ExpiresAt:    time.Now().Add(time.Hour),
		})

		start := time.Now()
		_, err := authSvc.Refresh(ctx, RefreshRequest{RefreshToken: plainToken})
		samples[i] = time.Since(start)
		if err != nil {
			t.Fatalf("Refresh failed: %v", err)
		}
	}

	got := p95(samples)
	want := 200 * time.Millisecond
	if got > want {
		t.Errorf("Refresh p95 latency = %v, want <= %v (SC-002)", got, want)
	}
}
