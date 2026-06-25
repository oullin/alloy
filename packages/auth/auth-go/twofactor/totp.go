package twofactor

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPeriod = 30
	defaultDigits = 6
)

// GenerateSecret creates a base32 TOTP secret.
func GenerateSecret() (string, error) {
	b := make([]byte, 20)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), "="), nil
}

// Code returns the current 6-digit TOTP code for a secret.
func Code(secret string, at time.Time) (string, error) {
	return code(secret, at, defaultDigits, defaultPeriod)
}

// Verify reports whether codeValue is valid within the given time-step window.
func Verify(secret, codeValue string, at time.Time, window int) bool {
	codeValue = strings.TrimSpace(codeValue)

	if codeValue == "" {
		return false
	}

	for offset := -window; offset <= window; offset++ {
		candidate, err := code(secret, at.Add(time.Duration(offset*defaultPeriod)*time.Second), defaultDigits, defaultPeriod)

		if err != nil {
			return false
		}

		if hmac.Equal([]byte(candidate), []byte(codeValue)) {
			return true
		}
	}

	return false
}

// OTPAuthURL returns an otpauth:// URL for authenticator apps.
func OTPAuthURL(issuer, account, secret string) string {
	label := issuer + ":" + account
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", issuer)
	values.Set("period", strconv.Itoa(defaultPeriod))
	values.Set("digits", strconv.Itoa(defaultDigits))
	values.Set("algorithm", "SHA1")

	return "otpauth://totp/" + url.PathEscape(label) + "?" + values.Encode()
}

func code(secret string, at time.Time, digits, period int) (string, error) {
	key, err := decodeSecret(secret)

	if err != nil {
		return "", err
	}

	counter := uint64(at.Unix() / int64(period))

	var buffer [8]byte

	binary.BigEndian.PutUint64(buffer[:], counter)

	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buffer[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	modulo := uint32(math.Pow10(digits))

	return fmt.Sprintf("%0*d", digits, truncated%modulo), nil
}

func decodeSecret(secret string) ([]byte, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))

	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(normalized)
}
