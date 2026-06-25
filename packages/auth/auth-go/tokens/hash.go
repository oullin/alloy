package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// GenerateSecret returns a URL-safe random token secret.
func GenerateSecret() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashSecret hashes a personal access token secret for storage.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))

	return hex.EncodeToString(sum[:])
}

// MatchSecret compares a stored secret hash with a plaintext secret.
func MatchSecret(storedHash, secret string) bool {
	if storedHash == "" || secret == "" {
		return false
	}

	candidate := HashSecret(secret)

	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(candidate)) == 1
}

func parsePlainTextToken(token string) (string, string, bool) {
	id, secret, ok := strings.Cut(strings.TrimSpace(token), "|")

	return id, secret, ok && id != "" && secret != ""
}
