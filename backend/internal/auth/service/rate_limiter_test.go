package service

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRateLimiter_Allow(t *testing.T) {
	cfg := RateLimiterConfig{
		MaxRequests: 3,
		Window:      time.Second,
		KeyPrefix:   "test:",
	}

	rl := NewMemoryRateLimiter(cfg)
	ctx := context.Background()
	key := "test-key"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		allowed, err := rl.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Allow() error: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	allowed, err := rl.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow() error: %v", err)
	}
	if allowed {
		t.Error("4th request should be denied")
	}

	// Different key should be allowed
	allowed, err = rl.Allow(ctx, "different-key")
	if err != nil {
		t.Fatalf("Allow() error: %v", err)
	}
	if !allowed {
		t.Error("Different key should be allowed")
	}
}

func TestMemoryRateLimiter_Reset(t *testing.T) {
	cfg := RateLimiterConfig{
		MaxRequests: 2,
		Window:      time.Minute,
		KeyPrefix:   "test:",
	}

	rl := NewMemoryRateLimiter(cfg)
	ctx := context.Background()
	key := "reset-key"

	// Use up the limit
	rl.Allow(ctx, key)
	rl.Allow(ctx, key)

	allowed, _ := rl.Allow(ctx, key)
	if allowed {
		t.Error("Should be rate limited")
	}

	// Reset
	if err := rl.Reset(ctx, key); err != nil {
		t.Fatalf("Reset() error: %v", err)
	}

	// Should be allowed again
	allowed, _ = rl.Allow(ctx, key)
	if !allowed {
		t.Error("Should be allowed after reset")
	}
}

func TestMemoryRateLimiter_GetRemaining(t *testing.T) {
	cfg := RateLimiterConfig{
		MaxRequests: 5,
		Window:      time.Minute,
		KeyPrefix:   "test:",
	}

	rl := NewMemoryRateLimiter(cfg)
	ctx := context.Background()
	key := "remaining-key"

	// Initially should have max remaining
	remaining, err := rl.GetRemaining(ctx, key)
	if err != nil {
		t.Fatalf("GetRemaining() error: %v", err)
	}
	if remaining != 5 {
		t.Errorf("GetRemaining() = %d, want 5", remaining)
	}

	// After 2 requests
	rl.Allow(ctx, key)
	rl.Allow(ctx, key)

	remaining, _ = rl.GetRemaining(ctx, key)
	if remaining != 3 {
		t.Errorf("GetRemaining() = %d, want 3", remaining)
	}
}

func TestMemoryRateLimiter_GetResetTime(t *testing.T) {
	cfg := RateLimiterConfig{
		MaxRequests: 5,
		Window:      time.Minute,
		KeyPrefix:   "test:",
	}

	rl := NewMemoryRateLimiter(cfg)
	ctx := context.Background()
	key := "time-key"

	// Before any requests
	resetTime, err := rl.GetResetTime(ctx, key)
	if err != nil {
		t.Fatalf("GetResetTime() error: %v", err)
	}

	// Make a request
	rl.Allow(ctx, key)

	resetTime, _ = rl.GetResetTime(ctx, key)

	// Reset time should be in the future
	if resetTime.Before(time.Now()) {
		t.Error("Reset time should be in the future")
	}

	// Reset time should be within the window
	if resetTime.After(time.Now().Add(cfg.Window + time.Second)) {
		t.Error("Reset time should be within window")
	}
}

func TestMemoryRateLimiter_SlidingWindow(t *testing.T) {
	cfg := RateLimiterConfig{
		MaxRequests: 2,
		Window:      100 * time.Millisecond,
		KeyPrefix:   "test:",
	}

	rl := NewMemoryRateLimiter(cfg)
	ctx := context.Background()
	key := "sliding-key"

	// Use up limit
	rl.Allow(ctx, key)
	rl.Allow(ctx, key)

	allowed, _ := rl.Allow(ctx, key)
	if allowed {
		t.Error("Should be rate limited")
	}

	// Wait for window to pass
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	allowed, _ = rl.Allow(ctx, key)
	if !allowed {
		t.Error("Should be allowed after window passes")
	}
}

func TestDefaultLoginRateLimiterConfig(t *testing.T) {
	cfg := DefaultLoginRateLimiterConfig()

	if cfg.MaxRequests != 5 {
		t.Errorf("MaxRequests = %d, want 5", cfg.MaxRequests)
	}
	if cfg.Window != time.Minute {
		t.Errorf("Window = %v, want 1 minute", cfg.Window)
	}
	if cfg.KeyPrefix != "login:" {
		t.Errorf("KeyPrefix = %q, want %q", cfg.KeyPrefix, "login:")
	}
}

func TestDefaultPasswordResetRateLimiterConfig(t *testing.T) {
	cfg := DefaultPasswordResetRateLimiterConfig()

	if cfg.MaxRequests != 1 {
		t.Errorf("MaxRequests = %d, want 1", cfg.MaxRequests)
	}
	if cfg.Window != 5*time.Minute {
		t.Errorf("Window = %v, want 5 minutes", cfg.Window)
	}
	if cfg.KeyPrefix != "password_reset:" {
		t.Errorf("KeyPrefix = %q, want %q", cfg.KeyPrefix, "password_reset:")
	}
}
