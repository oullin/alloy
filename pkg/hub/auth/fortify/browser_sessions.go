package fortify

import (
	"net/http"
	"time"

	"hara.sh/alloy/auth/browserx"
	cauth "hara.sh/alloy/contracts/auth"
)

type browserSessionResponse struct {
	ID           string    `json:"id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	LastActiveAt time.Time `json:"last_active_at"`
	Current      bool      `json:"current"`
}

// NewListBrowserSessionsHandler lists the current user's browser sessions.
func NewListBrowserSessionsHandler(guard cauth.Guard, service *browserx.Service, current CurrentSessionID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		sessions, err := service.List(r.Context(), user.GetAuthIdentifier(), currentSessionID(r, current))

		if err != nil {
			writeError(w, http.StatusInternalServerError, "browser sessions unavailable")

			return
		}

		payload := make([]browserSessionResponse, 0, len(sessions))

		for _, session := range sessions {
			payload = append(payload, browserSessionResponse{
				ID:           session.ID,
				IPAddress:    session.IPAddress,
				UserAgent:    session.UserAgent,
				LastActiveAt: session.LastActiveAt,
				Current:      session.Current,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{"sessions": payload})
	}
}

// NewRevokeBrowserSessionHandler revokes a single browser session.
func NewRevokeBrowserSessionHandler(guard cauth.Guard, service *browserx.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		id := r.PathValue("session")

		if id == "" {
			input, err := readInput(r)

			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid request")

				return
			}

			id = stringInput(input, "session")
		}

		if id == "" {
			writeValidation(w, "session", "required")

			return
		}

		if err := service.Revoke(r.Context(), user.GetAuthIdentifier(), id); err != nil {
			writeError(w, http.StatusNotFound, "browser session not found")

			return
		}

		writeNoContent(w)
	}
}

// NewRevokeOtherBrowserSessionsHandler revokes every other browser session.
func NewRevokeOtherBrowserSessionsHandler(guard cauth.Guard, service *browserx.Service, current CurrentSessionID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		if err := service.RevokeOther(r.Context(), user.GetAuthIdentifier(), currentSessionID(r, current)); err != nil {
			writeError(w, http.StatusInternalServerError, "browser sessions unavailable")

			return
		}

		writeNoContent(w)
	}
}

func authenticatedUser(w http.ResponseWriter, r *http.Request, guard cauth.Guard) (cauth.User, bool) {
	user, err := guard.User(r.Context())

	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")

		return nil, false
	}

	return user, true
}

func currentSessionID(r *http.Request, current CurrentSessionID) string {
	if current == nil {
		return ""
	}

	return current(r)
}
