package passkeys

import (
	"context"
	"net/http"

	"hara.sh/alloy/auth/fortify"
	cauth "hara.sh/alloy/contracts/auth"
)

// FortifyService adapts Service to fortify.PasskeyService without making the
// parent foundation module import WebAuthn.
type FortifyService struct {
	service *Service
}

// NewFortifyService creates a Fortify passkey service adapter.
func NewFortifyService(service *Service) FortifyService {
	return FortifyService{service: service}
}

func (s FortifyService) BeginRegistration(ctx context.Context, key string, user cauth.User) (any, error) {
	return s.service.BeginRegistration(ctx, key, user)
}

func (s FortifyService) FinishRegistration(ctx context.Context, key string, user cauth.User, r *http.Request) (any, error) {
	return s.service.FinishRegistration(ctx, key, user, r)
}

func (s FortifyService) BeginDiscoverableLogin(ctx context.Context, key string) (any, error) {
	return s.service.BeginDiscoverableLogin(ctx, key)
}

func (s FortifyService) FinishPasskeyLogin(ctx context.Context, key string, r *http.Request, resolveUser fortify.PasskeyUserResolver) (cauth.User, any, error) {
	return s.service.FinishPasskeyLogin(ctx, key, r, resolveUser)
}

var _ fortify.PasskeyService = FortifyService{}
