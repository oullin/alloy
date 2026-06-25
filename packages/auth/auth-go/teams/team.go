package teams

import "time"

// Team represents a user-owned team.
type Team struct {
	ID        string
	Name      string
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Member represents a user's membership in a team.
type Member struct {
	TeamID string
	UserID string
	Role   string
}

// Role describes a team role and its permissions.
type Role struct {
	Name        string
	Permissions []string
}

// Can reports whether the role grants permission.
func (r Role) Can(permission string) bool {
	for _, candidate := range r.Permissions {
		if candidate == "*" || candidate == permission {
			return true
		}
	}

	return false
}
