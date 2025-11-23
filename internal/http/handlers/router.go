package apihandler

import (
	"net/http"
	"time"

	"pr-reviewer-service/internal/storage/repo"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(store *repo.Store) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // на проде укажите домены
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	})
	r.Use(corsMiddleware.Handler)

	h := NewHandlers(store)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", h.CreateTeam)
		r.Get("/get", h.GetTeam)
		r.Post("/deactivate", h.BulkDeactivate)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", h.SetIsActive)
		r.Get("/getReview", h.GetPRsForReviewer)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", h.CreatePR)
		r.Post("/merge", h.MergePR)
		r.Post("/reassign", h.Reassign)
	})

	r.Route("/stats", func(r chi.Router) {
		r.Get("/reviewers", h.GetReviewerStats)
	})

	return r
}
