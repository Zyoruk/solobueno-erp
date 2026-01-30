package service

import (
	"context"
	"sync"
	"time"
)

// MemoryRateLimiter is an in-memory implementation of RateLimiter.
// Uses sliding window algorithm for accurate rate limiting.
type MemoryRateLimiter struct {
	config  RateLimiterConfig
	windows map[string]*slidingWindow
	mu      sync.RWMutex
}

// slidingWindow tracks requests in a sliding time window.
type slidingWindow struct {
	timestamps []time.Time
	mu         sync.Mutex
}

// NewMemoryRateLimiter creates a new in-memory rate limiter.
func NewMemoryRateLimiter(config RateLimiterConfig) *MemoryRateLimiter {
	rl := &MemoryRateLimiter{
		config:  config,
		windows: make(map[string]*slidingWindow),
	}

	// Start background cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given key should be allowed.
func (r *MemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := r.config.KeyPrefix + key
	now := time.Now()

	r.mu.Lock()
	window, exists := r.windows[fullKey]
	if !exists {
		window = &slidingWindow{
			timestamps: make([]time.Time, 0, r.config.MaxRequests),
		}
		r.windows[fullKey] = window
	}
	r.mu.Unlock()

	window.mu.Lock()
	defer window.mu.Unlock()

	// Remove timestamps outside the window
	cutoff := now.Add(-r.config.Window)
	validTimestamps := make([]time.Time, 0, len(window.timestamps))
	for _, ts := range window.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	window.timestamps = validTimestamps

	// Check if we're at the limit
	if len(window.timestamps) >= r.config.MaxRequests {
		return false, nil
	}

	// Add the current request
	window.timestamps = append(window.timestamps, now)
	return true, nil
}

// Reset resets the rate limit counter for the given key.
func (r *MemoryRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := r.config.KeyPrefix + key

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.windows, fullKey)
	return nil
}

// GetRemaining returns the number of remaining requests for the given key.
func (r *MemoryRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	fullKey := r.config.KeyPrefix + key
	now := time.Now()

	r.mu.RLock()
	window, exists := r.windows[fullKey]
	r.mu.RUnlock()

	if !exists {
		return r.config.MaxRequests, nil
	}

	window.mu.Lock()
	defer window.mu.Unlock()

	// Count valid timestamps
	cutoff := now.Add(-r.config.Window)
	count := 0
	for _, ts := range window.timestamps {
		if ts.After(cutoff) {
			count++
		}
	}

	remaining := r.config.MaxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// GetResetTime returns when the rate limit will reset for the given key.
func (r *MemoryRateLimiter) GetResetTime(ctx context.Context, key string) (time.Time, error) {
	fullKey := r.config.KeyPrefix + key
	now := time.Now()

	r.mu.RLock()
	window, exists := r.windows[fullKey]
	r.mu.RUnlock()

	if !exists || len(window.timestamps) == 0 {
		return now, nil
	}

	window.mu.Lock()
	defer window.mu.Unlock()

	// Find the oldest timestamp in the window
	cutoff := now.Add(-r.config.Window)
	var oldest time.Time
	for _, ts := range window.timestamps {
		if ts.After(cutoff) {
			if oldest.IsZero() || ts.Before(oldest) {
				oldest = ts
			}
		}
	}

	if oldest.IsZero() {
		return now, nil
	}

	// Reset time is when the oldest timestamp expires
	return oldest.Add(r.config.Window), nil
}

// cleanup periodically removes expired windows to prevent memory leaks.
func (r *MemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(r.config.Window)
	defer ticker.Stop()

	for range ticker.C {
		r.cleanupExpired()
	}
}

// cleanupExpired removes windows with no recent timestamps.
func (r *MemoryRateLimiter) cleanupExpired() {
	now := time.Now()
	cutoff := now.Add(-r.config.Window)

	r.mu.Lock()
	defer r.mu.Unlock()

	for key, window := range r.windows {
		window.mu.Lock()

		// Check if all timestamps are expired
		hasValid := false
		for _, ts := range window.timestamps {
			if ts.After(cutoff) {
				hasValid = true
				break
			}
		}

		window.mu.Unlock()

		if !hasValid {
			delete(r.windows, key)
		}
	}
}

// Ensure MemoryRateLimiter implements RateLimiter
var _ RateLimiter = (*MemoryRateLimiter)(nil)
