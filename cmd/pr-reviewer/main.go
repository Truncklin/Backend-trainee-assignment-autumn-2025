package main

import (
	"log/slog"
	"os"

	"pr-reviewer-service/internal/config"
	"pr-reviewer-service/internal/storage"

	//delete this line if no more imports are needed
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	//delete this line if no more code is needed
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("No .env file found, relying on environment variables")
	}

	cfg := config.MustLoadConfig()

	log := setupLogger(cfg.Env)

	storage, err := storage.NewPool(cfg.StoragePath)
	if err != nil {
		log.Error("StoragePath is incorrect", cfg.StoragePath)
	}
	_ = storage

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
