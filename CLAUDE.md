# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Requires `.env` (copy from `.env.example`) with `DB_CONN`, `CONFIG_PATH` (defaults to `./local/local.yaml`), `MIDDLEWARE_TOKEN`, and a reachable Postgres instance.

```bash
# Build
go build -o server.exe ./cmd/main
go build -o cli.exe ./cmd/cli

# Run (server must be started before the CLI)
./server.exe
./cli.exe

# Test
go test ./...                                    # all tests
go test ./internal/handlers/redirect/...          # one package
go test ./internal/handlers/alias/... -run TestAlias_DBError -v   # one test

# Vet (no golangci-lint config in this repo; go vet + gofmt are the available checks)
go vet ./...
gofmt -l .

# Docker (Postgres + server)
docker compose up --build
docker compose down          # -v to also drop the postgres_data volume
```

The server listens on the address in `local/local.yaml` (`http_server.addr`, default `0.0.0.0:8082`). Set it to `0.0.0.0` for Docker, `localhost` for bare-metal local runs.

## Architecture

Two independent binaries built from `cmd/main` (HTTP server) and `cmd/cli` (interactive terminal client). The CLI is just an HTTP client of the server's API (base URL hardcoded to `http://localhost:8082` in `cmd/cli/main.go`) — it has no direct DB or storage access.

**Request flow (server):** `cmd/main/main.go` wires a chi router with middleware applied in this order: `CustomMiddleware.NewLogger` → `CustomMiddleware.RateLimit` → `chi/middleware.Recoverer`. `POST /url` and `GET /{alias}` are public; `DELETE /url/{alias}` is registered in a chi sub-group with `CustomMiddleware.Auth` applied only there.

**Dependency inversion at the handler boundary:** each handler package declares its own single-method interface for exactly what it needs from storage (`alias.URLSaver`, `redirect.URLGetter`, `deleter.URLDeleter`) rather than depending on `*storage.Storage` directly. This is what makes the handlers mockable in tests without a real DB — see `internal/handlers/redirect/redirect_test.go` and `internal/handlers/deleter/deleter_test.go` for the pattern (a struct with a function field implementing the interface). Follow this same pattern for any new handler.

**Response shape is not fully unified.** `internal/handlers/api` exposes `JSONHandler(w, error, status, statusCode)` writing `{"error":..., "status":...}`, and `redirect`/`deleter` use it for all responses. `alias`, however, builds its own richer `Response{OriginalUrl, Alias, RespStatus{Status, Error}}` and does not use `api.JSONHandler`. Match whichever response convention the handler you're touching already uses rather than trying to unify them as a side effect.

**Storage (`internal/storage/storage.go`):** thin wrapper around `database/sql` + `lib/pq`, no ORM, no migration tool. `storage.New` opens the DB from `DB_CONN` and runs `CREATE TABLE IF NOT EXISTS` itself on every startup — there is no separate migrations step to keep in sync when changing the schema. `GetUrl` does the click-count increment and the read in one `UPDATE ... RETURNING` statement to stay atomic under concurrent redirects.

**Config (`internal/config/config.go`):** `godotenv.Load()` for `.env`, then `cleanenv.ReadConfig` for the YAML at `CONFIG_PATH`. Missing config file is a hard `log.Panic`, not a returned error.

**Auth (`internal/middleware/Auth.go`):** compares `Authorization: Bearer <token>` against `MIDDLEWARE_TOKEN`, read fresh from `.env` via `godotenv.Load()` on every request (not cached at startup).

**Rate limiting (`internal/middleware/ratelimiter.go`):** in-process map keyed by `r.RemoteAddr` (includes port), rejecting requests <500ms apart per key, with a background goroutine evicting entries idle >5min. This state is per-process/per-instance — it will not coordinate across multiple server replicas.

**Logging:** `internal/logger.SetupLogger(env)` switches handler by `cfg.Env` (`local` → text/debug, `test` → JSON/debug, `prod` → JSON/info, default → text/debug). Handlers derive request-scoped loggers via `log.With(slog.String("op", "handlers.X.New"))`; the custom logger middleware additionally logs method/path/status/duration/request-id per request.

**Alias generation (`lib/random/random.go`):** uses `math/rand`, not `crypto/rand` — intentional/acknowledged tradeoff for short-lived, low-security aliases, not an oversight to "fix".

## Code style notes

- Comments are written in Russian and are used to explain *why*, not *what* (e.g. rationale for atomic SQL, tradeoffs like `math/rand` choice) — match this when adding comments in existing files.
- Error wrapping uses `fmt.Errorf("...: %w", err)` throughout `internal/storage`.
- Package name for all middleware files is `CustomMiddleware`, not `middleware` (avoids collision with `github.com/go-chi/chi/v5/middleware`, which is also imported directly in `cmd/main/main.go` and `internal/middleware/logger.go`).
