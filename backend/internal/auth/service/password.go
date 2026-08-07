// Package service provides business logic services for the auth module.
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"unicode"

	"github.com/alexedwards/argon2id"
)

var (
	// ErrPasswordTooShort is returned when the password is too short.
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrPasswordTooWeak is returned when the password doesn't meet complexity requirements.
	ErrPasswordTooWeak = errors.New("password must contain at least one uppercase letter, one lowercase letter, and one number")
	// ErrPasswordMismatch is returned when password verification fails.
	ErrPasswordMismatch = errors.New("password does not match")
)

// PasswordService handles password hashing and verification.
type PasswordService struct {
	params *argon2id.Params
}

// NewPasswordService creates a new PasswordService with recommended OWASP parameters.
func NewPasswordService() *PasswordService {
	return &PasswordService{
		params: &argon2id.Params{
			Memory:      64 * 1024, // 64 MB
			Iterations:  3,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

// Hash creates a hash of the given password using Argon2id.
func (s *PasswordService) Hash(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, s.params)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return hash, nil
}

// Verify checks if the given password matches the hash.
func (s *PasswordService) Verify(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("failed to verify password: %w", err)
	}
	return match, nil
}

// ValidatePassword checks if a password meets the minimum requirements.
// Requirements:
// - At least 8 characters long
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one digit
func (s *PasswordService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrPasswordTooWeak
	}

	return nil
}

// GenerateTemporaryPassword generates a secure temporary password.
// The password will be 12 characters and meet all complexity requirements.
func (s *PasswordService) GenerateTemporaryPassword() (string, error) {
	const (
		upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowerChars = "abcdefghijklmnopqrstuvwxyz"
		digitChars = "0123456789"
		allChars   = upperChars + lowerChars + digitChars
		length     = 12
	)

	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Ensure at least one of each required character type
	password := make([]byte, length)
	password[0] = upperChars[int(bytes[0])%len(upperChars)]
	password[1] = lowerChars[int(bytes[1])%len(lowerChars)]
	password[2] = digitChars[int(bytes[2])%len(digitChars)]

	// Fill the rest randomly
	for i := 3; i < length; i++ {
		password[i] = allChars[int(bytes[i])%len(allChars)]
	}

	// Shuffle the password
	for i := len(password) - 1; i > 0; i-- {
		j := int(bytes[i]) % (i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

// GenerateResetToken generates a cryptographically secure reset token.
// Returns both the plain token (to send to user) and the hash (to store in DB).
func (s *PasswordService) GenerateResetToken() (plainToken, tokenHash string, err error) {
	// Generate 32 random bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64url for the plain token
	plainToken = base64.URLEncoding.EncodeToString(tokenBytes)

	// Create SHA256 hash for storage
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash = base64.URLEncoding.EncodeToString(hash[:])

	return plainToken, tokenHash, nil
}

// HashResetToken creates a hash of a plain reset token.
// Used to look up tokens in the database.
func (s *PasswordService) HashResetToken(plainToken string) string {
	hash := sha256.Sum256([]byte(plainToken))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// HashRefreshToken creates a hash of a plain refresh token.
// Refresh tokens are stored hashed in the database.
func (s *PasswordService) HashRefreshToken(plainToken string) string {
	hash := sha256.Sum256([]byte(plainToken))
	return base64.URLEncoding.EncodeToString(hash[:])
}
