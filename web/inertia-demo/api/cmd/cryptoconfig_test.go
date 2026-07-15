package main

import (
	"encoding/base64"
	"testing"
)

func TestLoadCrypto_FromEnv(t *testing.T) {
	envKey := base64.StdEncoding.EncodeToString(make([]byte, 32))

	t.Setenv(CryptoKeyEnvVar, envKey)

	cfg, err := LoadCrypto()

	if err != nil {
		t.Fatal(err)
	}

	if cfg.Key != envKey {
		t.Errorf("Key = %q, want %q", cfg.Key, envKey)
	}
}

func TestLoadCrypto_MissingEnvFailsFast(t *testing.T) {
	t.Setenv(CryptoKeyEnvVar, "")

	_, err := LoadCrypto()

	if err == nil {
		t.Fatal("expected error when crypto key env var is unset")
	}
}

func TestLoadCrypto_TrimsWhitespace(t *testing.T) {
	envKey := base64.StdEncoding.EncodeToString(make([]byte, 32))

	t.Setenv(CryptoKeyEnvVar, "  "+envKey+"  ")

	cfg, err := LoadCrypto()

	if err != nil {
		t.Fatal(err)
	}

	if cfg.Key != envKey {
		t.Errorf("Key = %q, want %q (whitespace trimmed)", cfg.Key, envKey)
	}
}

func TestCryptoConfig_DecodedKey(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 32)

	for i := range raw {
		raw[i] = byte(i)
	}

	cfg := CryptoConfig{
		Key: base64.StdEncoding.EncodeToString(raw),
	}

	key, err := cfg.DecodedKey()

	if err != nil {
		t.Fatal(err)
	}

	if len(key) != 32 {
		t.Errorf("key length = %d, want 32", len(key))
	}

	for i, b := range key {
		if b != byte(i) {
			t.Errorf("key[%d] = %d, want %d", i, b, i)
		}
	}
}

func TestCryptoConfig_DecodedKey_Empty(t *testing.T) {
	t.Parallel()

	cfg := CryptoConfig{}

	_, err := cfg.DecodedKey()

	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestCryptoConfig_DecodedKey_InvalidBase64(t *testing.T) {
	t.Parallel()

	cfg := CryptoConfig{Key: "not-valid-base64!!!"}

	_, err := cfg.DecodedKey()

	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestCryptoConfig_DecodedKey_WrongLength(t *testing.T) {
	t.Parallel()

	cfg := CryptoConfig{
		Key: base64.StdEncoding.EncodeToString(make([]byte, 16)),
	}

	_, err := cfg.DecodedKey()

	if err == nil {
		t.Error("expected error for 16-byte key (need 32)")
	}
}

func TestCryptoConfig_DecodedKey_Base64Prefix(t *testing.T) {
	t.Parallel()

	raw := make([]byte, 32)
	cfg := CryptoConfig{
		Key: "base64:" + base64.StdEncoding.EncodeToString(raw),
	}

	key, err := cfg.DecodedKey()

	if err != nil {
		t.Fatal(err)
	}

	if len(key) != 32 {
		t.Fatalf("key length = %d, want 32", len(key))
	}
}
