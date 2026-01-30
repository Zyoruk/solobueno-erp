// Package jwt provides JWT token generation and validation utilities.
package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"
)

var (
	// ErrKeyNotFound is returned when a key file cannot be found.
	ErrKeyNotFound = errors.New("key file not found")
	// ErrKeyInvalid is returned when a key cannot be parsed.
	ErrKeyInvalid = errors.New("key is invalid")
	// ErrKeyNotLoaded is returned when attempting to use a key that hasn't been loaded.
	ErrKeyNotLoaded = errors.New("key has not been loaded")
)

// KeyManager handles RSA key loading and caching for JWT operations.
type KeyManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	mu         sync.RWMutex
}

// NewKeyManager creates a new KeyManager instance.
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// LoadPrivateKeyFromFile loads an RSA private key from a PEM file.
func (km *KeyManager) LoadPrivateKeyFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrKeyNotFound, path)
		}
		return fmt.Errorf("failed to read key file: %w", err)
	}

	return km.LoadPrivateKeyFromPEM(data)
}

// LoadPrivateKeyFromPEM loads an RSA private key from PEM data.
func (km *KeyManager) LoadPrivateKeyFromPEM(pemData []byte) error {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return fmt.Errorf("%w: failed to decode PEM block", ErrKeyInvalid)
	}

	var privateKey *rsa.PrivateKey
	var err error

	// Try PKCS#8 first (modern format)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("%w: not an RSA private key", ErrKeyInvalid)
		}
	} else {
		// Fall back to PKCS#1 (traditional RSA format)
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrKeyInvalid, err)
		}
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	km.privateKey = privateKey
	km.publicKey = &privateKey.PublicKey

	return nil
}

// LoadPublicKeyFromFile loads an RSA public key from a PEM file.
func (km *KeyManager) LoadPublicKeyFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrKeyNotFound, path)
		}
		return fmt.Errorf("failed to read key file: %w", err)
	}

	return km.LoadPublicKeyFromPEM(data)
}

// LoadPublicKeyFromPEM loads an RSA public key from PEM data.
func (km *KeyManager) LoadPublicKeyFromPEM(pemData []byte) error {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return fmt.Errorf("%w: failed to decode PEM block", ErrKeyInvalid)
	}

	var publicKey *rsa.PublicKey

	// Try PKIX format first
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		var ok bool
		publicKey, ok = key.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("%w: not an RSA public key", ErrKeyInvalid)
		}
	} else {
		// Try PKCS#1 format
		publicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrKeyInvalid, err)
		}
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	km.publicKey = publicKey

	return nil
}

// LoadKeysFromEnv loads keys from environment variables.
// Expects JWT_PRIVATE_KEY and JWT_PUBLIC_KEY to contain PEM-encoded keys,
// or JWT_PRIVATE_KEY_FILE and JWT_PUBLIC_KEY_FILE to contain file paths.
func (km *KeyManager) LoadKeysFromEnv() error {
	// Try direct PEM content first
	if privateKeyPEM := os.Getenv("JWT_PRIVATE_KEY"); privateKeyPEM != "" {
		if err := km.LoadPrivateKeyFromPEM([]byte(privateKeyPEM)); err != nil {
			return fmt.Errorf("failed to load private key from env: %w", err)
		}
	} else if privateKeyFile := os.Getenv("JWT_PRIVATE_KEY_FILE"); privateKeyFile != "" {
		if err := km.LoadPrivateKeyFromFile(privateKeyFile); err != nil {
			return fmt.Errorf("failed to load private key from file: %w", err)
		}
	}

	if publicKeyPEM := os.Getenv("JWT_PUBLIC_KEY"); publicKeyPEM != "" {
		if err := km.LoadPublicKeyFromPEM([]byte(publicKeyPEM)); err != nil {
			return fmt.Errorf("failed to load public key from env: %w", err)
		}
	} else if publicKeyFile := os.Getenv("JWT_PUBLIC_KEY_FILE"); publicKeyFile != "" {
		if err := km.LoadPublicKeyFromFile(publicKeyFile); err != nil {
			return fmt.Errorf("failed to load public key from file: %w", err)
		}
	}

	// Load key ID from env
	km.keyID = os.Getenv("JWT_KEY_ID")
	if km.keyID == "" {
		km.keyID = "key-1" // Default key ID
	}

	return nil
}

// SetKeyID sets the key ID used in JWT headers.
func (km *KeyManager) SetKeyID(keyID string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.keyID = keyID
}

// GetKeyID returns the current key ID.
func (km *KeyManager) GetKeyID() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.keyID
}

// GetPrivateKey returns the loaded private key.
func (km *KeyManager) GetPrivateKey() (*rsa.PrivateKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.privateKey == nil {
		return nil, ErrKeyNotLoaded
	}
	return km.privateKey, nil
}

// GetPublicKey returns the loaded public key.
func (km *KeyManager) GetPublicKey() (*rsa.PublicKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.publicKey == nil {
		return nil, ErrKeyNotLoaded
	}
	return km.publicKey, nil
}

// HasPrivateKey returns true if a private key has been loaded.
func (km *KeyManager) HasPrivateKey() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.privateKey != nil
}

// HasPublicKey returns true if a public key has been loaded.
func (km *KeyManager) HasPublicKey() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.publicKey != nil
}
