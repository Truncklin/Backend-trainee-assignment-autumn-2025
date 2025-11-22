package main

import (
	"log/slog"
	"os"

	"pr-reviewer-service/internal/config"
	"pr-reviewer-service/internal/lib/logger"
	"pr-reviewer-service/internal/storage"
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

	// TODO: init router: chi

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
