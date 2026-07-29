package tokens

import (
	"context"
	"time"

	cauth "hara.sh/alloy/contracts/auth"
)

// Issuer creates personal access tokens for authenticated users.
type Issuer struct {
	repo Repository
}

// NewIssuer creates an Issuer.
func NewIssuer(repo Repository) *Issuer {
	return &Issuer{repo: repo}
}

// CreateToken creates a hashed personal access token and returns its plaintext
// form once.
func (i *Issuer) CreateToken(ctx context.Context, user cauth.User, name string, abilities []string, expiresAt *time.Time) (PlainTextToken, error) {
	secret, err := GenerateSecret()

	if err != nil {
		return PlainTextToken{}, err
	}

	token, err := i.repo.Create(ctx, Token{
		UserID:    user.GetAuthIdentifier(),
		Name:      name,
		TokenHash: HashSecret(secret),
		Abilities: normalizeAbilities(abilities),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	})

	if err != nil {
		return PlainTextToken{}, err
	}

	return PlainTextToken{
		AccessToken: token,
		PlainText:   token.ID + "|" + secret,
	}, nil
}

func normalizeAbilities(abilities []string) []string {
	if len(abilities) == 0 {
		return []string{"*"}
	}

	normalized := make([]string, 0, len(abilities))

	for _, ability := range abilities {
		if ability != "" {
			normalized = append(normalized, ability)
		}
	}

	if len(normalized) == 0 {
		return []string{"*"}
	}

	return normalized
}
