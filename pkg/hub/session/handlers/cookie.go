package handlers

import (
	"context"
	"net/http"
)

// CookieEncrypter encrypts and decrypts session cookie payloads.
type CookieEncrypter interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// CookieHandler stores the entire session payload in a signed/encrypted HTTP
// cookie. It requires a request/response pair to be injected per-request via
// SetRequest/SetWriter before Read/Write are called.
type CookieHandler struct {
	enc     CookieEncrypter
	name    string
	request *http.Request
	writer  http.ResponseWriter
	// secure is a tri-state Secure flag for the emitted cookie: nil (the
	// default) means secure-by-default (true); an explicit false set via
	// SetSecure is honored so local HTTP dev can opt out.
	secure *bool
}

// NewCookieHandler creates a CookieHandler. name is the cookie name. The
// session cookie is emitted with Secure=true by default; call SetSecure(false)
// to opt out for local HTTP development.
func NewCookieHandler(enc CookieEncrypter, name string) *CookieHandler {
	return &CookieHandler{enc: enc, name: name}
}

// SetRequest sets the incoming request for cookie reading.
func (h *CookieHandler) SetRequest(r *http.Request) { h.request = r }

// SetWriter sets the response writer for cookie writing.
func (h *CookieHandler) SetWriter(w http.ResponseWriter) { h.writer = w }

// SetSecure overrides the Secure flag on emitted cookies. Unset, the handler
// defaults to Secure=true; pass false to opt out for local HTTP development.
func (h *CookieHandler) SetSecure(secure bool) { h.secure = &secure }

// isSecure resolves the tri-state: nil defaults to true.
func (h *CookieHandler) isSecure() bool { return h.secure == nil || *h.secure }

func (h *CookieHandler) Open(_ context.Context, _, _ string) error { return nil }

func (h *CookieHandler) Close(_ context.Context) error { return nil }

func (h *CookieHandler) Read(_ context.Context, _ string) (string, error) {
	if h.request == nil {
		return "", nil
	}

	c, err := h.request.Cookie(h.name)

	if err != nil {
		return "", nil
	}

	plaintext, err := h.enc.Decrypt(c.Value)

	if err != nil {
		return "", nil
	}

	return plaintext, nil
}

func (h *CookieHandler) Write(_ context.Context, _, data string) error {
	if h.writer == nil {
		return nil
	}

	ciphertext, err := h.enc.Encrypt(data)

	if err != nil {
		return err
	}

	http.SetCookie(h.writer, &http.Cookie{
		Name:     h.name,
		Value:    ciphertext,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (h *CookieHandler) Destroy(_ context.Context, _ string) error {
	if h.writer == nil {
		return nil
	}

	http.SetCookie(h.writer, &http.Cookie{
		Name:   h.name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
		Secure: h.isSecure(),
	})

	return nil
}

func (h *CookieHandler) GC(_ context.Context, _ int) error { return nil }
