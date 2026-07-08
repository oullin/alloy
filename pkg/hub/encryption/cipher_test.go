package encryption

import (
	"errors"
	"testing"
)

func TestParseCipherAllSupported(t *testing.T) {
	t.Parallel()

	cases := map[string]Cipher{
		"aes-128-cbc": AES128CBC,
		"aes-256-cbc": AES256CBC,
		"aes-128-gcm": AES128GCM,
		"aes-256-gcm": AES256GCM,
	}

	for name, want := range cases {
		c, err := ParseCipher(name)

		if err != nil {
			t.Fatalf("ParseCipher(%q) failed: %v", name, err)
		}

		if c != want {
			t.Fatalf("expected %s, got %s", want, c)
		}
	}
}

func TestKeyLengthPerCipher(t *testing.T) {
	t.Parallel()

	cases := map[Cipher]int{
		AES128CBC:        16,
		AES128GCM:        16,
		AES256CBC:        32,
		AES256GCM:        32,
		Cipher("bogus"):  0,
		Cipher(""):       0,
		Cipher("aes-xy"): 0,
	}

	for c, want := range cases {
		if got := c.KeyLength(); got != want {
			t.Fatalf("KeyLength(%s) = %d, want %d", c, got, want)
		}
	}
}

func TestIVLengthPerCipher(t *testing.T) {
	t.Parallel()

	cases := map[Cipher]int{
		AES128CBC:       16,
		AES256CBC:       16,
		AES128GCM:       12,
		AES256GCM:       12,
		Cipher("bogus"): 0,
	}

	for c, want := range cases {
		if got := c.IVLength(); got != want {
			t.Fatalf("IVLength(%s) = %d, want %d", c, got, want)
		}
	}
}

func TestIsAEAD(t *testing.T) {
	t.Parallel()

	if AES128CBC.IsAEAD() || AES256CBC.IsAEAD() {
		t.Fatal("CBC ciphers should not be AEAD")
	}

	if !AES128GCM.IsAEAD() || !AES256GCM.IsAEAD() {
		t.Fatal("GCM ciphers should be AEAD")
	}
}

func TestParseKeyRawBytes(t *testing.T) {
	t.Parallel()

	key, err := ParseKey("plain-key-material")

	if err != nil {
		t.Fatal(err)
	}

	if string(key) != "plain-key-material" {
		t.Fatalf("expected raw bytes back, got %q", key)
	}
}

func TestParseKeyEmptyString(t *testing.T) {
	t.Parallel()

	_, err := ParseKey("")

	if !errors.Is(err, ErrUnsupportedCipher) {
		t.Fatalf("expected ErrUnsupportedCipher, got %v", err)
	}
}

func TestParseKeyInvalidBase64(t *testing.T) {
	t.Parallel()

	_, err := ParseKey("base64:!!!not-base64!!!")

	if !errors.Is(err, ErrUnsupportedCipher) {
		t.Fatalf("expected ErrUnsupportedCipher, got %v", err)
	}
}

func TestGenerateKeyUnsupportedCipher(t *testing.T) {
	t.Parallel()

	_, err := GenerateKey(Cipher("aes-512-xyz"))

	if !errors.Is(err, ErrUnsupportedCipher) {
		t.Fatalf("expected ErrUnsupportedCipher, got %v", err)
	}
}

func TestSupportedRejectsUnknownCipher(t *testing.T) {
	t.Parallel()

	if Supported(make([]byte, 16), Cipher("bogus")) {
		t.Fatal("unknown cipher should never be supported")
	}
}
