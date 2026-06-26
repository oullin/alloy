package fortify

import (
	"net/http"
	"time"

	patokens "github.com/oullin/alloy/api/auth/tokens"
	cauth "github.com/oullin/alloy/api/contracts/auth"
)

type apiTokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Abilities  []string   `json:"abilities"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// NewListAPITokensHandler returns personal access tokens for the current user.
func NewListAPITokensHandler(guard cauth.Guard, repo patokens.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := guard.User(r.Context())

		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")

			return
		}

		found, err := repo.FindForUser(r.Context(), user.GetAuthIdentifier())

		if err != nil {
			writeError(w, http.StatusInternalServerError, "api tokens unavailable")

			return
		}

		tokens := make([]apiTokenResponse, 0, len(found))

		for _, token := range found {
			tokens = append(tokens, tokenResponse(token))
		}

		writeJSON(w, http.StatusOK, map[string]any{"tokens": tokens})
	}
}

// NewCreateAPITokenHandler creates a hashed personal access token.
func NewCreateAPITokenHandler(guard cauth.Guard, issuer *patokens.Issuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := guard.User(r.Context())

		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")

			return
		}

		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		name := stringInput(input, "name")

		if name == "" {
			writeValidation(w, "name", "required")

			return
		}

		expiresAt, err := expiresAtInput(input)

		if err != nil {
			writeValidation(w, "expires_at", "must be RFC3339")

			return
		}

		created, err := issuer.CreateToken(r.Context(), user, name, abilitiesInput(input), expiresAt)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "api token creation failed")

			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"token":      tokenResponse(created.AccessToken),
			"plain_text": created.PlainText,
		})
	}
}

// NewRevokeAPITokenHandler revokes a personal access token owned by the user.
func NewRevokeAPITokenHandler(guard cauth.Guard, repo patokens.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := guard.User(r.Context())

		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")

			return
		}

		id := r.PathValue("token")

		if id == "" {
			input, err := readInput(r)

			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid request")

				return
			}

			id = stringInput(input, "token")
		}

		if id == "" {
			writeValidation(w, "token", "required")

			return
		}

		if err := repo.Revoke(r.Context(), id, user.GetAuthIdentifier()); err != nil {
			writeError(w, http.StatusNotFound, "api token not found")

			return
		}

		writeNoContent(w)
	}
}

func tokenResponse(token patokens.Token) apiTokenResponse {
	return apiTokenResponse{
		ID:         token.ID,
		Name:       token.Name,
		Abilities:  append([]string(nil), token.Abilities...),
		CreatedAt:  token.CreatedAt,
		LastUsedAt: token.LastUsedAt,
		ExpiresAt:  token.ExpiresAt,
		RevokedAt:  token.RevokedAt,
	}
}

func abilitiesInput(input map[string]any) []string {
	value, ok := input["abilities"]

	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		abilities := make([]string, 0, len(typed))

		for _, item := range typed {
			if ability, ok := item.(string); ok && ability != "" {
				abilities = append(abilities, ability)
			}
		}

		return abilities
	case string:
		if typed == "" {
			return nil
		}

		return []string{typed}
	default:
		return nil
	}
}

func expiresAtInput(input map[string]any) (*time.Time, error) {
	value := stringInput(input, "expires_at")

	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)

	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
