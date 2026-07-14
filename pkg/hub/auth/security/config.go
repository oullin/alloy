package security

import (
	"errors"
	"fmt"
	"time"
)

// Config captures production security defaults for the headless auth stack.
type Config struct {
	AppKey                string
	SecureCookies         bool
	HTTPOnlyCookies       bool
	CSRFVerifyOrigin      bool
	SessionLifetime       time.Duration
	PasswordResetExpiry   time.Duration
	PasswordResetThrottle time.Duration
	LoginMaxAttempts      int
	LoginDecay            time.Duration
	LoginLockout          time.Duration
	PersonalTokenExpiry   time.Duration
	Passkeys              PasskeyConfig
}

// PasskeyConfig captures WebAuthn relying-party defaults.
type PasskeyConfig struct {
	RPID          string
	RPDisplayName string
	RPOrigins     []string
}

var (
	ErrMissingAppKey        = errors.New("security: app key is required")
	ErrMissingPasskeyRPID   = errors.New("security: passkey relying party id is required")
	ErrMissingPasskeyOrigin = errors.New("security: passkey origin is required")
)

// ProductionDefaults returns conservative defaults for browser-backed auth.
func ProductionDefaults() Config {
	return Config{
		SecureCookies:         true,
		HTTPOnlyCookies:       true,
		CSRFVerifyOrigin:      true,
		SessionLifetime:       2 * time.Hour,
		PasswordResetExpiry:   time.Hour,
		PasswordResetThrottle: time.Minute,
		LoginMaxAttempts:      5,
		LoginDecay:            time.Minute,
		LoginLockout:          time.Minute,
		PersonalTokenExpiry:   365 * 24 * time.Hour,
	}
}

// ValidateProduction rejects missing or unsafe production security settings.
func (c Config) ValidateProduction() error {
	if c.AppKey == "" {
		return ErrMissingAppKey
	}

	if !c.SecureCookies {
		return fmt.Errorf("security: secure cookies must be enabled")
	}

	if !c.HTTPOnlyCookies {
		return fmt.Errorf("security: http-only cookies must be enabled")
	}

	if !c.CSRFVerifyOrigin {
		return fmt.Errorf("security: csrf origin verification must be enabled")
	}

	if c.SessionLifetime <= 0 {
		return fmt.Errorf("security: session lifetime must be positive")
	}

	if c.PasswordResetExpiry <= 0 {
		return fmt.Errorf("security: password reset expiry must be positive")
	}

	if c.PasswordResetThrottle <= 0 {
		return fmt.Errorf("security: password reset throttle must be positive")
	}

	if c.LoginMaxAttempts <= 0 || c.LoginDecay <= 0 || c.LoginLockout <= 0 {
		return fmt.Errorf("security: login throttling must be configured")
	}

	if c.Passkeys.RPID == "" {
		return ErrMissingPasskeyRPID
	}

	if len(c.Passkeys.RPOrigins) == 0 {
		return ErrMissingPasskeyOrigin
	}

	return nil
}
