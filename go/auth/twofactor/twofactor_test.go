package twofactor_test

import (
	"strings"
	"testing"
	"time"

	"alloy.dev/go/auth/twofactor"
)

func TestCodeAndVerifyTOTP(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	at := time.Unix(59, 0)

	code, err := twofactor.Code(secret, at)

	if err != nil {
		t.Fatal(err)
	}

	if code != "287082" {
		t.Fatalf("code = %q, want 287082", code)
	}

	if !twofactor.Verify(secret, code, at, 0) {
		t.Fatal("expected code to verify at exact time")
	}

	if !twofactor.Verify(secret, code, at.Add(30*time.Second), 1) {
		t.Fatal("expected code to verify inside window")
	}

	if twofactor.Verify(secret, "000000", at, 1) {
		t.Fatal("expected wrong code to fail")
	}
}

func TestGenerateSecretAndOTPAuthURL(t *testing.T) {
	secret, err := twofactor.GenerateSecret()

	if err != nil {
		t.Fatal(err)
	}

	if secret == "" {
		t.Fatal("expected secret")
	}

	url := twofactor.OTPAuthURL("Alloy", "taylor@example.com", secret)

	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Fatalf("url = %q", url)
	}

	if !strings.Contains(url, "issuer=Alloy") {
		t.Fatalf("url missing issuer: %q", url)
	}
}

func TestRecoveryCodesAreHashedAndSingleUse(t *testing.T) {
	codes, err := twofactor.GenerateRecoveryCodes(2)

	if err != nil {
		t.Fatal(err)
	}

	if len(codes) != 2 {
		t.Fatalf("codes = %d, want 2", len(codes))
	}

	hashes := twofactor.HashRecoveryCodes(codes)

	if hashes[0] == codes[0] {
		t.Fatal("recovery code stored in plaintext")
	}

	remaining, ok := twofactor.UseRecoveryCode(hashes, codes[0])

	if !ok {
		t.Fatal("expected recovery code to be used")
	}

	if len(remaining) != 1 {
		t.Fatalf("remaining = %d, want 1", len(remaining))
	}

	if _, ok := twofactor.UseRecoveryCode(remaining, codes[0]); ok {
		t.Fatal("expected recovery code to be single-use")
	}
}
