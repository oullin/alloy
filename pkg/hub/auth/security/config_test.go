package security_test

import (
	"errors"
	"testing"

	"github.com/oullin/alloy/pkg/hub/auth/security"
)

func TestProductionDefaultsValidateWhenRequiredSecretsAreSet(t *testing.T) {
	cfg := security.ProductionDefaults()
	cfg.AppKey = "base64:test"
	cfg.Passkeys.RPID = "example.com"
	cfg.Passkeys.RPDisplayName = "Alloy"
	cfg.Passkeys.RPOrigins = []string{"https://example.com"}

	if err := cfg.ValidateProduction(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateProductionRejectsMissingAppKey(t *testing.T) {
	cfg := security.ProductionDefaults()
	cfg.Passkeys.RPID = "example.com"
	cfg.Passkeys.RPOrigins = []string{"https://example.com"}

	if err := cfg.ValidateProduction(); !errors.Is(err, security.ErrMissingAppKey) {
		t.Fatalf("err = %v, want ErrMissingAppKey", err)
	}
}

func TestValidateProductionRejectsUnsafeCookieAndPasskeyConfig(t *testing.T) {
	cfg := security.ProductionDefaults()
	cfg.AppKey = "base64:test"
	cfg.SecureCookies = false
	cfg.Passkeys.RPID = "example.com"
	cfg.Passkeys.RPOrigins = []string{"https://example.com"}

	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected insecure cookies to be rejected")
	}

	cfg = security.ProductionDefaults()
	cfg.AppKey = "base64:test"

	if err := cfg.ValidateProduction(); !errors.Is(err, security.ErrMissingPasskeyRPID) {
		t.Fatalf("err = %v, want ErrMissingPasskeyRPID", err)
	}
}
