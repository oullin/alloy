package passkeys

import (
	"context"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	cauth "github.com/oullin/alloy/auth/contracts/auth"
)

// Service coordinates WebAuthn ceremonies.
type Service struct {
	webauthn *webauthn.WebAuthn
	repo     Repository
	sessions SessionStore
}

// NewService creates a WebAuthn passkey service.
func NewService(config *webauthn.Config, repo Repository, sessions SessionStore) (*Service, error) {
	wa, err := webauthn.New(config)

	if err != nil {
		return nil, err
	}

	return &Service{webauthn: wa, repo: repo, sessions: sessions}, nil
}

// BeginRegistration starts a passkey registration ceremony.
func (s *Service) BeginRegistration(ctx context.Context, key string, user cauth.Authenticatable) (*protocol.CredentialCreation, error) {
	waUser, err := s.user(ctx, user)

	if err != nil {
		return nil, err
	}

	options, session, err := s.webauthn.BeginRegistration(waUser)

	if err != nil {
		return nil, err
	}

	if err := s.sessions.Put(ctx, key, *session); err != nil {
		return nil, err
	}

	return options, nil
}

// FinishRegistration validates the response and stores the credential.
func (s *Service) FinishRegistration(ctx context.Context, key string, user cauth.Authenticatable, r *http.Request) (*webauthn.Credential, error) {
	waUser, err := s.user(ctx, user)

	if err != nil {
		return nil, err
	}

	session, err := s.sessions.Get(ctx, key)

	if err != nil {
		return nil, err
	}

	credential, err := s.webauthn.FinishRegistration(waUser, session, r)

	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveCredential(ctx, user.GetAuthIdentifier(), *credential); err != nil {
		return nil, err
	}

	_ = s.sessions.Delete(ctx, key)

	return credential, nil
}

// BeginDiscoverableLogin starts a passkey login ceremony.
func (s *Service) BeginDiscoverableLogin(ctx context.Context, key string) (*protocol.CredentialAssertion, error) {
	options, session, err := s.webauthn.BeginDiscoverableLogin()

	if err != nil {
		return nil, err
	}

	if err := s.sessions.Put(ctx, key, *session); err != nil {
		return nil, err
	}

	return options, nil
}

// FinishPasskeyLogin validates a discoverable passkey login response.
func (s *Service) FinishPasskeyLogin(ctx context.Context, key string, r *http.Request, resolveUser func(context.Context, string) (cauth.Authenticatable, error)) (cauth.Authenticatable, *webauthn.Credential, error) {
	session, err := s.sessions.Get(ctx, key)

	if err != nil {
		return nil, nil, err
	}

	var authenticated cauth.Authenticatable
	user, credential, err := s.webauthn.FinishPasskeyLogin(func(_, userHandle []byte) (webauthn.User, error) {
		userID, err := s.repo.UserIDByHandle(ctx, userHandle)

		if err != nil {
			return nil, err
		}

		authUser, err := resolveUser(ctx, userID)

		if err != nil {
			return nil, err
		}

		authenticated = authUser

		return s.user(ctx, authUser)
	}, session, r)

	if err != nil {
		return nil, nil, err
	}

	if authenticated == nil {
		if adapted, ok := user.(User); ok {
			authenticated = adapted.Auth
		}
	}

	if authenticated != nil && credential != nil {
		_ = s.repo.UpdateCredential(ctx, authenticated.GetAuthIdentifier(), *credential)
	}

	_ = s.sessions.Delete(ctx, key)

	return authenticated, credential, nil
}

func (s *Service) user(ctx context.Context, user cauth.Authenticatable) (User, error) {
	userID := user.GetAuthIdentifier()
	handle, err := s.repo.GetOrCreateUserHandle(ctx, userID)

	if err != nil {
		return User{}, err
	}

	credentials, err := s.repo.CredentialsByUser(ctx, userID)

	if err != nil {
		return User{}, err
	}

	return User{Auth: user, Handle: handle, Credentials: credentials}, nil
}
