package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// CryptoConfig holds the encryption key used for cookie encryption.
// The key must be a base64-encoded 32-byte value for AES-256-CBC.
type CryptoConfig struct {
	Key string
}

// CryptoKeyEnvVar is the environment variable that supplies the base64-encoded
// AES-256 key used for cookie encryption, CSRF, and the flash store.
//
// The key is a secret and must never be committed to the repository. Generate a
// value with a CSPRNG, for example:
//
//	openssl rand -base64 32
const CryptoKeyEnvVar = "INERTIA_CRYPTO_KEY"

// LoadCrypto reads the encryption key from the environment.
//
// The key is a secret and is supplied only through the CryptoKeyEnvVar
// environment variable; it is never read from a committed file. A missing key
// is a fatal misconfiguration and returns an error so that startup fails fast
// rather than falling back to a built-in default.
func LoadCrypto() (CryptoConfig, error) {
	key := strings.TrimSpace(os.Getenv(CryptoKeyEnvVar))

	if key == "" {
		return CryptoConfig{}, fmt.Errorf(
			"crypto: %s environment variable is required; generate one with `openssl rand -base64 32`",
			CryptoKeyEnvVar,
		)
	}

	return CryptoConfig{Key: key}, nil
}

// DecodedKey returns the raw 32-byte AES-256 key from the base64-encoded
// Key field. It returns an error if the key is missing, not valid base64,
// or not exactly 32 bytes.
func (c *CryptoConfig) DecodedKey() ([]byte, error) {
	if strings.TrimSpace(c.Key) == "" {
		return nil, fmt.Errorf("crypto: key is required")
	}

	encoded := strings.TrimSpace(c.Key)
	encoded = strings.TrimPrefix(encoded, "base64:")

	key, err := base64.StdEncoding.DecodeString(encoded)

	if err != nil {
		return nil, fmt.Errorf("crypto: invalid base64 key: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("crypto: key must be 32 bytes, got %d", len(key))
	}

	return key, nil
}
