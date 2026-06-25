package session_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/oullin/alloy/session"
	"github.com/oullin/alloy/session/handlers"
)

// fakeEncrypter applies base64 encoding as a reversible transformation.
type fakeEncrypter struct{}

type failingEncrypter struct {
	encryptErr error
	decryptErr error
}

func (e *fakeEncrypter) Encrypt(plaintext string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(plaintext)), nil
}

func (e *fakeEncrypter) Decrypt(ciphertext string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(ciphertext)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (e *failingEncrypter) Encrypt(_ string) (string, error) {
	if e.encryptErr != nil {
		return "", e.encryptErr
	}

	return "encrypted", nil
}

func (e *failingEncrypter) Decrypt(_ string) (string, error) {
	if e.decryptErr != nil {
		return "", e.decryptErr
	}

	return "plain", nil
}

func TestEncryptedStorePutGetString(t *testing.T) {
	h := handlers.NewArrayHandler()
	enc := &fakeEncrypter{}
	es := session.NewEncryptedStore("test", h, enc)

	_ = es.Start(context.Background())
	es.Put("secret", "hello")

	// The encrypted store should return the plaintext.
	if got := es.Get("secret", nil); got != "hello" {
		t.Errorf("expected hello, got %v", got)
	}

	// The inner store should hold the encrypted (base64) value.
	raw := es.Store.Get("secret", nil)
	expected := base64.StdEncoding.EncodeToString([]byte("hello"))

	if raw != expected {
		t.Errorf("inner store should hold encrypted value: got %v, want %v", raw, expected)
	}
}

func TestEncryptedStoreStartSave(t *testing.T) {
	h := handlers.NewArrayHandler()
	enc := &fakeEncrypter{}
	es := session.NewEncryptedStore("test", h, enc)

	if err := es.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	es.Put("key", "value")

	if err := es.Save(context.Background()); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load into a new encrypted store and verify round-trip.
	es2 := session.NewEncryptedStore("test", h, enc)
	_ = es2.Store.SetID(es.Store.GetID())
	_ = es2.Start(context.Background())

	if got := es2.Get("key", nil); got != "value" {
		t.Errorf("expected value after round-trip, got %v", got)
	}
}

func TestEncryptedStorePutNonString(t *testing.T) {
	h := handlers.NewArrayHandler()
	enc := &fakeEncrypter{}
	es := session.NewEncryptedStore("test", h, enc)

	_ = es.Start(context.Background())
	es.Put("num", 42)

	// Non-string values pass through unencrypted.
	if got := es.Get("num", nil); got != 42 {
		t.Errorf("expected 42, got %v", got)
	}
}

func TestEncryptedStoreDoesNotStorePlaintextWhenEncryptionFails(t *testing.T) {
	h := handlers.NewArrayHandler()
	es := session.NewEncryptedStore("test", h, &failingEncrypter{encryptErr: errors.New("encrypt failed")})

	_ = es.Start(context.Background())
	es.Put("secret", "plain")

	if es.Store.Exists("secret") {
		t.Fatal("expected failed encryption to leave key unset")
	}
}

func TestEncryptedStoreReturnsFallbackWhenDecryptFails(t *testing.T) {
	h := handlers.NewArrayHandler()
	es := session.NewEncryptedStore("test", h, &failingEncrypter{decryptErr: errors.New("decrypt failed")})

	_ = es.Start(context.Background())
	es.Store.Put("secret", "ciphertext")

	if got := es.Get("secret", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback on decrypt failure, got %v", got)
	}
}
