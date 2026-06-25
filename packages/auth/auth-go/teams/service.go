package teams

import "context"

// Service coordinates team workflows.
type Service struct {
	repo  Repository
	roles map[string]Role
}

// NewService creates a Service.
func NewService(repo Repository, roles []Role) *Service {
	byName := make(map[string]Role, len(roles))

	for _, role := range roles {
		byName[role.Name] = role
	}

	return &Service{repo: repo, roles: byName}
}

func (s *Service) Create(ctx context.Context, ownerID, name string) (Team, error) {
	return s.repo.CreateTeam(ctx, Team{Name: name, OwnerID: ownerID})
}

func (s *Service) List(ctx context.Context, userID string) ([]Team, error) {
	return s.repo.TeamsForUser(ctx, userID)
}

func (s *Service) AddMember(ctx context.Context, actorID, teamID, userID, role string) error {
	if err := s.authorize(ctx, actorID, teamID, "members:create"); err != nil {
		return err
	}

	return s.repo.AddMember(ctx, Member{TeamID: teamID, UserID: userID, Role: role})
}

func (s *Service) UpdateRole(ctx context.Context, actorID, teamID, userID, role string) error {
	if err := s.authorize(ctx, actorID, teamID, "members:update"); err != nil {
		return err
	}

	return s.repo.UpdateMemberRole(ctx, teamID, userID, role)
}

func (s *Service) RemoveMember(ctx context.Context, actorID, teamID, userID string) error {
	if err := s.authorize(ctx, actorID, teamID, "members:delete"); err != nil {
		return err
	}

	return s.repo.RemoveMember(ctx, teamID, userID)
}

func (s *Service) SwitchCurrent(ctx context.Context, userID, teamID string) error {
	return s.repo.SetCurrentTeam(ctx, userID, teamID)
}

func (s *Service) Current(ctx context.Context, userID string) (Team, error) {
	return s.repo.CurrentTeam(ctx, userID)
}

func (s *Service) authorize(ctx context.Context, actorID, teamID, permission string) error {
	member, err := s.repo.Member(ctx, teamID, actorID)

	if err != nil {
		return ErrForbidden
	}

	if member.Role == "owner" {
		return nil
	}

	role, ok := s.roles[member.Role]

	if !ok || !role.Can(permission) {
		return ErrForbidden
	}

	return nil
}
