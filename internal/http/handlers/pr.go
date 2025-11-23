package apihandler

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/storage/repo"
)

func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PRID   string `json:"pull_request_id"`
		Name   string `json:"pull_request_name"`
		Author string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "BAD_REQUEST", "invalid json")
		return
	}

	pr, err := h.store.CreatePR(r.Context(), body.PRID, body.Name, body.Author)
	switch err {
	case repo.ErrNotFound:
		writeError(w, 404, "NOT_FOUND", "author not found")
		return
	case repo.ErrPRExists:
		writeError(w, 409, "PR_EXISTS", "pr already exists")
		return
	}

	writeJSON(w, 201, map[string]any{"pr": pr})
}

func (h *Handlers) MergePR(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PRID string `json:"pull_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "BAD_REQUEST", "invalid json")
		return
	}

	pr, err := h.store.MergePR(r.Context(), body.PRID)
	if err == repo.ErrNotFound {
		writeError(w, 404, "NOT_FOUND", "pr not found")
		return
	}

	writeJSON(w, 200, map[string]any{"pr": pr})
}

func (h *Handlers) Reassign(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PRID    string `json:"pull_request_id"`
		OldUser string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "BAD_REQUEST", "invalid json")
		return
	}

	newID, err := h.store.ReassignReviewer(r.Context(), body.PRID, body.OldUser)
	switch err {
	case repo.ErrPRMerged:
		writeError(w, 409, "PR_MERGED", "cannot reassign merged PR")
		return
	case repo.ErrNotAssigned:
		writeError(w, 409, "NOT_ASSIGNED", "reviewer not assigned")
		return
	case repo.ErrNoCandidate:
		writeError(w, 409, "NO_CANDIDATE", "no candidate found")
		return
	case repo.ErrNotFound:
		writeError(w, 404, "NOT_FOUND", "pr or user not found")
		return
	}

	pr, _ := h.store.MergePR(r.Context(), body.PRID) // получить полные данные PR

	writeJSON(w, 200, map[string]any{
		"pr":          pr,
		"replaced_by": newID,
	})
}
