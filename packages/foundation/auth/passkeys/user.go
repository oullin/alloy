package passkeys

import (
	cauth "alloy.dev/foundation/contracts/auth"
	"github.com/go-webauthn/webauthn/webauthn"
)

// User adapts Alloy auth users to the WebAuthn user interface.
type User struct {
	Auth        cauth.User
	Handle      []byte
	Credentials []webauthn.Credential
}

func (u User) WebAuthnID() []byte {
	return append([]byte(nil), u.Handle...)
}

func (u User) WebAuthnName() string {
	return u.Auth.GetAuthIdentifier()
}

func (u User) WebAuthnDisplayName() string {
	return u.Auth.GetAuthIdentifier()
}

func (u User) WebAuthnCredentials() []webauthn.Credential {
	return append([]webauthn.Credential(nil), u.Credentials...)
}
