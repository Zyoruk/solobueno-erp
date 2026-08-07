package service

import (
	"context"
	"time"
)

// RateLimiter defines the interface for rate limiting operations.
type RateLimiter interface {
	// Allow checks if a request from the given key should be allowed.
	// Returns true if allowed, false if rate limit exceeded.
	Allow(ctx context.Context, key string) (bool, error)

	// Reset resets the rate limit counter for the given key.
	Reset(ctx context.Context, key string) error

	// GetRemaining returns the number of remaining requests for the given key.
	GetRemaining(ctx context.Context, key string) (int, error)

	// GetResetTime returns when the rate limit will reset for the given key.
	GetResetTime(ctx context.Context, key string) (time.Time, error)
}

// RateLimiterConfig holds configuration for rate limiters.
type RateLimiterConfig struct {
	// MaxRequests is the maximum number of requests allowed within the window.
	MaxRequests int

	// Window is the time window for rate limiting.
	Window time.Duration

	// KeyPrefix is the prefix to use for rate limiter keys.
	KeyPrefix string
}

// DefaultLoginRateLimiterConfig returns the default config for login rate limiting.
// Per FR-011: 5 requests per minute per IP.
func DefaultLoginRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxRequests: 5,
		Window:      time.Minute,
		KeyPrefix:   "login:",
	}
}

// DefaultPasswordResetRateLimiterConfig returns the default config for password reset rate limiting.
// 1 request per email per 5 minutes.
func DefaultPasswordResetRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxRequests: 1,
		Window:      5 * time.Minute,
		KeyPrefix:   "password_reset:",
	}
}
