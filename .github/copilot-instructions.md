# Copilot / AI Agent Instructions — pr-reviewer-service

Purpose: quick, actionable guidance so an AI coding assistant is immediately productive in this repository.

Key files
- `cmd/pr-reviewer/main.go`: single application entrypoint. Initialize config, logger, storage, router here.
- `config/local.yaml`: local configuration file (env, other keys live here).
- `go.mod`: Go module file — project targets `go 1.25.1`.

Big picture
- This repository is a small Go service with a single command under `cmd/pr-reviewer`.
- `main.go` is intentionally minimal and expects initialization steps for: config (cleanenv), logger (slog), storage (Postgres), and HTTP router (chi). Add higher-level application logic in non-`cmd` packages.

Project-specific conventions and patterns
- Entrypoint: put only wiring/initialization in `cmd/pr-reviewer/main.go`. Business logic, handlers, DB, and config types belong in packages under the repo root (create `internal/` or `pkg/` as needed).
- Config: prefer YAML files under `config/` for defaults and use environment variables for overrides (the code comments reference `cleanenv`). When adding config structs, ensure they are loaded early in `main` and passed explicitly to components.
- Logging: follow the TODO in `main.go` — project plans indicate `slog`. Initialize structured logger once and pass it to components; avoid package-level loggers.
- DB/storage: TODO mentions PostgreSQL. Use a repository/adapter pattern (create an `internal/storage` package) and keep SQL or ORM code isolated.
- Routing: `chi` is planned. Keep HTTP handlers in `internal/handlers` and wiring (routes -> handlers) in a small router package.

Build / run / test (PowerShell examples)
- Build the binary:
```powershell
go build -o .\bin\pr-reviewer ./cmd/pr-reviewer
```
- Run directly (useful during development):
```powershell
go run ./cmd/pr-reviewer --config=config/local.yaml
```
- Run all tests:
```powershell
go test ./... -v
```

What to look for when editing `main.go`
- Keep `main.go` focused on wiring. When you see a TODO (config/logger/storage/router), implement a constructor in an appropriate package and call it from `main`.
- Example flow to implement:
  1. load config from `config/local.yaml` + env
  2. init `slog` logger
  3. connect to PostgreSQL (provide a retry/backoff if needed)
  4. create router (chi), register handlers, and start HTTP server

Integration points and external dependencies
- The code comments indicate intent to use: `cleanenv`, `slog`, `chi`, and PostgreSQL. When adding dependencies, update `go.mod` via `go get` — keep modules tidy.

Editing guidance for AI agents
- Prefer small, testable changes. Add unit tests alongside new packages.
- When adding configuration keys, update `config/local.yaml` with sensible defaults.
- Avoid making broad repo layout changes without user confirmation (e.g., adding many top-level packages). Propose changes first.

If something is unclear
- Ask the developer where they want business logic and DB adapters to live (suggest `internal/` by default).
- Confirm desired libraries if you plan to implement TODOs (e.g., use `slog` vs `zap`).

Next steps (for humans):
- If you want, I can implement the TODOs in `cmd/pr-reviewer/main.go` and scaffold `internal/config`, `internal/logger`, `internal/storage`, and `internal/router` with minimal code and tests.
