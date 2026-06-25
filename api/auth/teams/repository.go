package teams

import (
	"context"
	"errors"
)

// Repository persists teams and memberships.
type Repository interface {
	CreateTeam(ctx context.Context, team Team) (Team, error)
	ForUser(ctx context.Context, userID string) ([]Team, error)
	FindTeam(ctx context.Context, teamID string) (Team, error)
	AddMember(ctx context.Context, member Member) error
	UpdateMemberRole(ctx context.Context, teamID, userID, role string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	Member(ctx context.Context, teamID, userID string) (Member, error)
	SetCurrentTeam(ctx context.Context, userID, teamID string) error
	CurrentTeam(ctx context.Context, userID string) (Team, error)
}

var (
	ErrTeamNotFound   = errors.New("teams: team not found")
	ErrMemberNotFound = errors.New("teams: member not found")
	ErrForbidden      = errors.New("teams: forbidden")
)
