package apihandler

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/storage/repo"
)

func (h *Handlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var t repo.Team
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, 400, "BAD_REQUEST", "invalid json")
		return
	}

	err := h.store.CreateTeam(r.Context(), t)
	if err != nil {
		writeError(w, 400, "TEAM_EXISTS", "team already exists")
		return
	}

	writeJSON(w, 201, map[string]any{"team": t})
}

func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("team_name")
	if name == "" {
		writeError(w, 400, "BAD_REQUEST", "team_name required")
		return
	}

	team, err := h.store.GetTeam(r.Context(), name)
	if err == repo.ErrNotFound {
		writeError(w, 404, "NOT_FOUND", "team not found")
		return
	} else if err != nil {
		writeError(w, 500, "INTERNAL", err.Error())
		return
	}

	writeJSON(w, 200, team)
}
