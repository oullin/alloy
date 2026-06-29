package tokens

import "time"

// Token is a hashed personal access token record.
type Token struct {
	ID         string
	UserID     string
	Name       string
	TokenHash  string
	Abilities  []string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	RevokedAt  *time.Time
}

// PlainTextToken is returned once, immediately after token creation.
type PlainTextToken struct {
	AccessToken Token
	PlainText   string
}

// Can reports whether the token grants an ability.
func (t Token) Can(ability string) bool {
	if t.RevokedAt != nil {
		return false
	}

	if t.ExpiresAt != nil && !t.ExpiresAt.After(time.Now()) {
		return false
	}

	for _, candidate := range t.Abilities {
		if candidate == "*" || candidate == ability {
			return true
		}
	}

	return false
}

// Cant reports whether the token does not grant an ability.
func (t Token) Cant(ability string) bool {
	return !t.Can(ability)
}

// Expired reports whether the token is past its expiry.
func (t Token) Expired(now time.Time) bool {
	return t.ExpiresAt != nil && !t.ExpiresAt.After(now)
}

// Revoked reports whether the token was revoked.
func (t Token) Revoked() bool {
	return t.RevokedAt != nil
}

// Active reports whether the token can currently be used.
func (t Token) Active(now time.Time) bool {
	return !t.Revoked() && !t.Expired(now)
}
