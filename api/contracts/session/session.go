package session

import (
	"context"
	"net/http"
)

// Handler abstracts the session storage backend.
type Handler interface {
	Open(ctx context.Context, path, name string) error
	Close(ctx context.Context) error
	Read(ctx context.Context, id string) (string, error)
	Write(ctx context.Context, id, data string) error
	Destroy(ctx context.Context, id string) error
	GC(ctx context.Context, maxLifetime int) error
}

// ExistenceAware is implemented by handlers that distinguish insert/update.
type ExistenceAware interface {
	SetExists(exists bool)
}

// Encrypter encrypts and decrypts session payloads.
type Encrypter interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// RequestAware is implemented by handlers that need the current HTTP request.
type RequestAware interface {
	SetRequest(r *http.Request)
}

// Cache is the minimal cache interface required by CacheHandler.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string, ttlSeconds int) error
	Forget(ctx context.Context, key string) error
}

// CSRFStore is the session surface required by VerifyCSRFToken.
type CSRFStore interface {
	Token() string
}

// VerifyCSRFConfig configures CSRF request verification.
type VerifyCSRFConfig struct {
	Except       []string
	HeaderName   string
	FormField    string
	VerifyOrigin bool
}
