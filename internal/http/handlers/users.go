package apihandler

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/storage/repo"
)

func (h *Handlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "BAD_REQUEST", "invalid json")
		return
	}

	user, err := h.store.SetUserActive(r.Context(), body.UserID, body.IsActive)
	if err == repo.ErrNotFound {
		writeError(w, 404, "NOT_FOUND", "user not found")
		return
	}

	writeJSON(w, 200, map[string]any{"user": user})
}

func (h *Handlers) GetPRsForReviewer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("user_id")
	if id == "" {
		writeError(w, 400, "BAD_REQUEST", "user_id required")
		return
	}

	prs, err := h.store.GetPRsForReviewer(r.Context(), id)
	if err != nil {
		writeError(w, 500, "INTERNAL", err.Error())
		return
	}

	writeJSON(w, 200, map[string]any{
		"user_id":       id,
		"pull_requests": prs,
	})
}
