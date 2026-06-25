package browsersessions

import (
	"context"
	"errors"
)

// Repository lists and revokes browser sessions.
type Repository interface {
	FindForUser(ctx context.Context, userID string) ([]Session, error)
	Revoke(ctx context.Context, userID, sessionID string) error
	RevokeOther(ctx context.Context, userID, currentSessionID string) error
}

// Service coordinates browser-session listing and revocation.
type Service struct {
	repo Repository
}

var ErrSessionNotFound = errors.New("browser sessions: session not found")

// NewService creates a Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns a user's sessions and marks the current session.
func (s *Service) List(ctx context.Context, userID, currentSessionID string) ([]Session, error) {
	sessions, err := s.repo.FindForUser(ctx, userID)

	if err != nil {
		return nil, err
	}

	for i := range sessions {
		sessions[i].Current = sessions[i].ID == currentSessionID
	}

	return sessions, nil
}

// Revoke revokes one session owned by the user.
func (s *Service) Revoke(ctx context.Context, userID, sessionID string) error {
	return s.repo.Revoke(ctx, userID, sessionID)
}

// RevokeOther revokes every session for the user except the current session.
func (s *Service) RevokeOther(ctx context.Context, userID, currentSessionID string) error {
	return s.repo.RevokeOther(ctx, userID, currentSessionID)
}
