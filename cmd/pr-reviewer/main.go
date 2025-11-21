package main

import (
	"log/slog"
	"os"
	"pr-reviewer-service/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	сfg := config.MustLoadConfig()

	log := setupLogger(сfg.Env)

	log.Info("starting pr-reviewer service", slog.String("env", сfg.Env))
	log.Debug("debug mode is enabled")

	// TODO: init storage: PostgreSQL

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
