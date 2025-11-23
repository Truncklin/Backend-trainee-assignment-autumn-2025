package apihandler

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/storage/repo"
)

type Handlers struct {
	store *repo.Store
}

func NewHandlers(s *repo.Store) *Handlers {
	return &Handlers{store: s}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}
