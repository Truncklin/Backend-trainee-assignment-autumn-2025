package apihandler

import (
	"net/http"
)

func (h *Handlers) GetReviewerStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetReviewerAssignmentStats(r.Context())
	if err != nil {
		writeError(w, 500, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"stats": stats})
}
