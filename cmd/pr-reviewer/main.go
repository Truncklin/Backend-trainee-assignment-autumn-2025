package main

import (
	"log/slog"
	"net/http"
	"os"

	"pr-reviewer-service/internal/config"
	apihandler "pr-reviewer-service/internal/http/handlers"
	"pr-reviewer-service/internal/lib/logger"
	"pr-reviewer-service/internal/storage"
	"pr-reviewer-service/internal/storage/repo"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoadConfig()

	log := setupLogger(cfg.Env)

	pool, err := storage.NewPool(cfg.StoragePath)
	if err != nil {
		log.Error("StoragePath is incorrect", slog.String("storage_path", cfg.StoragePath), logger.Err(err)) // slog.Attr{Key: "err",Value: slog.StringValue(err.Error()) == logger.Err(err)
		os.Exit(1)
	}

	err = storage.RunMigrations(pool)
	if err != nil {
		log.Error("Failed to run migrations", logger.Err(err))
		os.Exit(1)
	}

	store := repo.New(pool)
	r := apihandler.NewRouter(store)
	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	log.Info("starting server", slog.String("addr", cfg.HTTPServer.Address))
	if err := server.ListenAndServe(); err != nil {
		log.Error("server error", logger.Err(err))
	}

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
