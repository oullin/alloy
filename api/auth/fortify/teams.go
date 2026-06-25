package fortify

import (
	"net/http"
	"time"

	cauth "github.com/oullin/alloy/api/contracts/auth"
	"github.com/oullin/alloy/auth/teams"
)

type teamResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewListTeamsHandler(guard cauth.Guard, service *teams.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		found, err := service.List(r.Context(), user.GetAuthIdentifier())

		if err != nil {
			writeError(w, http.StatusInternalServerError, "teams unavailable")

			return
		}

		payload := make([]teamResponse, 0, len(found))

		for _, team := range found {
			payload = append(payload, teamPayload(team))
		}

		writeJSON(w, http.StatusOK, map[string]any{"teams": payload})
	}
}

func NewCreateTeamHandler(guard cauth.Guard, service *teams.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
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

		team, err := service.Create(r.Context(), user.GetAuthIdentifier(), name)

		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "team creation failed")

			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"team": teamPayload(team)})
	}
}

func NewSwitchCurrentTeamHandler(guard cauth.Guard, service *teams.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		teamID := stringInput(input, "team")

		if teamID == "" {
			writeValidation(w, "team", "required")

			return
		}

		if err := service.SwitchCurrent(r.Context(), user.GetAuthIdentifier(), teamID); err != nil {
			writeError(w, http.StatusForbidden, "team access denied")

			return
		}

		writeNoContent(w)
	}
}

func NewAddTeamMemberHandler(guard cauth.Guard, service *teams.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		teamID, targetUserID, role, ok := teamMemberInput(w, r, false)

		if !ok {
			return
		}

		if err := service.AddMember(r.Context(), user.GetAuthIdentifier(), teamID, targetUserID, role); err != nil {
			writeError(w, http.StatusForbidden, "team access denied")

			return
		}

		writeNoContent(w)
	}
}

func NewUpdateTeamMemberRoleHandler(guard cauth.Guard, service *teams.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		teamID, targetUserID, role, ok := teamMemberInput(w, r, true)

		if !ok {
			return
		}

		if err := service.UpdateRole(r.Context(), user.GetAuthIdentifier(), teamID, targetUserID, role); err != nil {
			writeError(w, http.StatusForbidden, "team access denied")

			return
		}

		writeNoContent(w)
	}
}

func NewRemoveTeamMemberHandler(guard cauth.Guard, service *teams.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		teamID, targetUserID, _, ok := teamMemberInput(w, r, true)

		if !ok {
			return
		}

		if err := service.RemoveMember(r.Context(), user.GetAuthIdentifier(), teamID, targetUserID); err != nil {
			writeError(w, http.StatusForbidden, "team access denied")

			return
		}

		writeNoContent(w)
	}
}

func teamPayload(team teams.Team) teamResponse {
	return teamResponse{
		ID:        team.ID,
		Name:      team.Name,
		OwnerID:   team.OwnerID,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
	}
}

func teamMemberInput(w http.ResponseWriter, r *http.Request, userFromPath bool) (string, string, string, bool) {
	input, err := readInput(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")

		return "", "", "", false
	}

	teamID := r.PathValue("team")

	if teamID == "" {
		teamID = stringInput(input, "team")
	}

	userID := ""

	if userFromPath {
		userID = r.PathValue("user")
	}

	if userID == "" {
		userID = stringInput(input, "user")
	}

	role := stringInput(input, "role")

	if role == "" {
		role = "member"
	}

	if teamID == "" {
		writeValidation(w, "team", "required")

		return "", "", "", false
	}

	if userID == "" {
		writeValidation(w, "user", "required")

		return "", "", "", false
	}

	return teamID, userID, role, true
}
