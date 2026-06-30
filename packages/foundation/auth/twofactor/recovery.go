package twofactor

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
)

// GenerateRecoveryCodes returns plaintext recovery codes for one-time display.
func GenerateRecoveryCodes(count int) ([]string, error) {
	if count <= 0 {
		count = 8
	}

	codes := make([]string, count)

	for i := range count {
		b := make([]byte, 10)

		if _, err := rand.Read(b); err != nil {
			return nil, err
		}

		codes[i] = base64.RawURLEncoding.EncodeToString(b)
	}

	return codes, nil
}

// HashRecoveryCode hashes a recovery code for storage.
func HashRecoveryCode(code string) string {
	sum := sha256.Sum256([]byte(code))

	return hex.EncodeToString(sum[:])
}

// HashRecoveryCodes hashes recovery codes for storage.
func HashRecoveryCodes(codes []string) []string {
	hashed := make([]string, 0, len(codes))

	for _, code := range codes {
		if code != "" {
			hashed = append(hashed, HashRecoveryCode(code))
		}
	}

	return hashed
}

// UseRecoveryCode removes a matching recovery code hash.
func UseRecoveryCode(storedHashes []string, code string) ([]string, bool) {
	if code == "" {
		return append([]string(nil), storedHashes...), false
	}

	candidate := HashRecoveryCode(code)
	remaining := make([]string, 0, len(storedHashes))
	used := false

	for _, stored := range storedHashes {
		if !used && subtle.ConstantTimeCompare([]byte(stored), []byte(candidate)) == 1 {
			used = true

			continue
		}

		remaining = append(remaining, stored)
	}

	return remaining, used
}
