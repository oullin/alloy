package teams

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// MemoryRepository stores teams in memory.
type MemoryRepository struct {
	mu          sync.RWMutex
	nextID      int64
	teams       map[string]Team
	members     map[string]map[string]Member
	currentTeam map[string]string
}

// NewMemoryRepository creates a MemoryRepository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		teams:       make(map[string]Team),
		members:     make(map[string]map[string]Member),
		currentTeam: make(map[string]string),
	}
}

func (r *MemoryRepository) CreateTeam(_ context.Context, team Team) (Team, error) {
	r.mu.Lock()

	defer r.mu.Unlock()

	r.nextID++
	team.ID = strconv.FormatInt(r.nextID, 10)
	now := time.Now()
	team.CreatedAt = now
	team.UpdatedAt = now
	r.teams[team.ID] = team

	if r.members[team.ID] == nil {
		r.members[team.ID] = make(map[string]Member)
	}

	r.members[team.ID][team.OwnerID] = Member{TeamID: team.ID, UserID: team.OwnerID, Role: "owner"}
	r.currentTeam[team.OwnerID] = team.ID

	return team, nil
}

func (r *MemoryRepository) TeamsForUser(_ context.Context, userID string) ([]Team, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	teams := make([]Team, 0)

	for teamID, members := range r.members {
		if _, ok := members[userID]; ok {
			teams = append(teams, r.teams[teamID])
		}
	}

	return teams, nil
}

func (r *MemoryRepository) FindTeam(_ context.Context, teamID string) (Team, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	team, ok := r.teams[teamID]

	if !ok {
		return Team{}, ErrTeamNotFound
	}

	return team, nil
}

func (r *MemoryRepository) AddMember(_ context.Context, member Member) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	if _, ok := r.teams[member.TeamID]; !ok {
		return ErrTeamNotFound
	}

	if r.members[member.TeamID] == nil {
		r.members[member.TeamID] = make(map[string]Member)
	}

	r.members[member.TeamID][member.UserID] = member

	return nil
}

func (r *MemoryRepository) UpdateMemberRole(_ context.Context, teamID, userID, role string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	member, ok := r.members[teamID][userID]

	if !ok {
		return ErrMemberNotFound
	}

	member.Role = role
	r.members[teamID][userID] = member

	return nil
}

func (r *MemoryRepository) RemoveMember(_ context.Context, teamID, userID string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	if _, ok := r.members[teamID][userID]; !ok {
		return ErrMemberNotFound
	}

	delete(r.members[teamID], userID)

	if r.currentTeam[userID] == teamID {
		delete(r.currentTeam, userID)
	}

	return nil
}

func (r *MemoryRepository) Member(_ context.Context, teamID, userID string) (Member, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	member, ok := r.members[teamID][userID]

	if !ok {
		return Member{}, ErrMemberNotFound
	}

	return member, nil
}

func (r *MemoryRepository) SetCurrentTeam(_ context.Context, userID, teamID string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	if _, ok := r.members[teamID][userID]; !ok {
		return ErrForbidden
	}

	r.currentTeam[userID] = teamID

	return nil
}

func (r *MemoryRepository) CurrentTeam(_ context.Context, userID string) (Team, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	teamID := r.currentTeam[userID]
	team, ok := r.teams[teamID]

	if !ok {
		return Team{}, ErrTeamNotFound
	}

	return team, nil
}
