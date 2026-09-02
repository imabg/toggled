# Feature 01 — Project Foundation

**Depends on:** nothing (build this first)
**Shared context:** see [`../plan.md`](../plan.md) for architecture, tech stack, and project layout.

## Goal

Stand up the Go project skeleton, local dev environment, and tooling that
every other feature builds on top of: module setup, config loading, a
Postgres connection pool, migration tooling, and a running HTTP server with
a health check.

## Deliverables

1. **Go module** at repo root (`go.mod`) targeting a current stable Go
   version.
2. **Project layout** per [`../plan.md`](../plan.md#project-layout):
   `cmd/server/main.go`, `internal/config`, `internal/api`, `internal/domain`,
   `internal/service`, `internal/store/postgres`, `internal/stream`,
   `db/migrations`.
3. **Config loading** (`internal/config`): reads from `config.yaml`. A
   committed `config.example.yaml` documents the schema; the real
   `config.yaml` is gitignored. Required config:
   -    `http.port`
   - `database.type` + typed backend block (v1: `postgres.url`)
   - `log.level`
   - `env` (`local` or `production`)
4. **Logger** (`uber-go/zap`): JSON structured output in production, a
   human-readable console encoder when `log.level` is `debug` or `env` is
   `local`. Logs go to stdout. `NewLogger` installs the logger as
   the process-wide default (`zap.L` / `zap.S`).
5. **DB connection**: a `pgxpool.Pool` constructed from `DATABASE_URL`,
   with a startup ping and sane defaults (max conns, connect timeout).
6. **HTTP server** (`cmd/server/main.go`): `gorilla/mux` router, wraps
   handlers with logging + panic-recovery middleware, graceful shutdown on
   `SIGINT`/`SIGTERM` (stop accepting new connections, let in-flight
   requests finish within a timeout, then close the DB pool).
7. **Health check endpoint**: `GET /healthz` → `200 {"status": "ok"}`. Should
   also verify DB connectivity (return `503` if the DB ping fails).
8. **Migrations tooling**: `golang-migrate` wired against `db/migrations`.
   No flag-related migrations yet — this feature just proves the tooling
   works (e.g. a trivial first migration, or leave `db/migrations` empty and
   let feature 02 add the first real migration).
9. **Local dev environment**:
   - `docker-compose.yml` with a `postgres` service (persistent named
     volume, exposed on `5432`, sane default credentials for local dev).
   - `Dockerfile` for the Go app (multi-stage build).
   - `Makefile` targets: `run`, `build`, `migrate-up`, `migrate-down`,
     `migrate-create name=...`, `test`, `lint`, `generate` (for `sqlc`).

## Out of scope

- No flags, no domain logic yet.
- No authentication/authorization — that's handled entirely by authx,
  outside this repo (see [`../plan.md`](../plan.md#scope-v1)).

## Acceptance criteria

- `docker-compose up -d postgres` starts a local Postgres instance.
- `make run` starts the HTTP server and logs a structured "server started"
  message with the listening port.
- `curl localhost:8080/healthz` returns `200 {"status":"ok"}` when Postgres
  is reachable, and a non-200 response when it isn't.
- `make migrate-up` / `make migrate-down` run against `DATABASE_URL`
  without error.
- Sending `SIGINT` to the running server logs a graceful shutdown sequence
  and exits cleanly (no dangling connections).
