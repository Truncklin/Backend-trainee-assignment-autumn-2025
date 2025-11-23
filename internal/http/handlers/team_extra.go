package apihandler

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/storage/repo"
)

func (h *Handlers) BulkDeactivate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TeamName string `json:"team_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "BAD_REQUEST", "invalid json")
		return
	}
	if body.TeamName == "" {
		writeError(w, 400, "BAD_REQUEST", "team_name required")
		return
	}

	res, err := h.store.BulkDeactivateTeam(r.Context(), body.TeamName)
	if err == repo.ErrNotFound {
		writeError(w, 404, "NOT_FOUND", "team not found")
		return
	}
	if err != nil {
		writeError(w, 500, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, 200, res)
}
