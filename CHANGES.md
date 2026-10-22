# Template Review: Summary of Changes

**Branch:** `refactor/template-review` (not committed yet)
**Goal:** A **bare**, framework-agnostic hexagonal Go skeleton that works when cloned as-is, where any framework (Gin, Echo, Fiber, gRPC, ...) or tool (Redis, MinIO, PostgreSQL, ...) can be plugged in as an adapter.

The work happened in two phases:
1. **Review & fix**: fixed the bugs in the original template and made it production-ready.
2. **Bare template**: removed the sample and all infrastructure from the base, and added a lifecycle module so anything can be plugged in. Recipes moved to `docs/recipes.md`.

---

## 1. Verification (final state)

| Check | Result |
|-------|--------|
| `go build ./...`, `go vet ./...` | ✅ Pass |
| `golangci-lint run` / `fmt --diff` | ✅ 0 issues / no diff |
| `go mod tidy -diff` | ✅ Tidy |
| `go test -race -count=3 ./...` | ✅ Pass (lifecycle 94%, config 88% coverage) |
| `make run` on the bare skeleton | ✅ Starts, warns "no servers registered", exits 0 |
| Recipes: net/http, Gin, Echo, Fiber, worker | ✅ Compiled, linted, **and run end to end** (requests served, 404 mapping, graceful shutdown) |
| Recipes: PostgreSQL (pgx), Redis, MinIO | ✅ Compiled and linted against the real libraries. ⚠️ **Not run against live servers** (Docker daemon wasn't running). |
| `docker build` | ⚠️ Not checked (Docker daemon wasn't running) |

---

## 2. Bugs fixed from the original template

| # | Problem | Fix |
|---|---------|-----|
| 1 | Following the README crashed the app: `APP_PORT=your-app-port // ...` can't be read as a number | `.env.example` now has valid values and `#` comments |
| 2 | Crashed when `.env` was missing, which is normal in Docker, Kubernetes and CI | `.env` is optional |
| 3 | `dev.yaml` / `prod.yaml` were never loaded (`config.Load("")`) | `config.Load(os.Getenv("APP_ENV"))` |
| 4 | The database config was ignored; `test.db` was hard-coded | The database is no longer in the base; recipes read it from config |
| 5 | The SQLite file could be committed by accident | `.gitignore` covers `*.db*` and `bin/` |
| 6 | `BuildContainer` returned `nil`, printed the service, and called `panic` | Replaced by `BuildApp` → `(*lifecycle.App, error)` |
| 7 | `go.mod` wasn't tidy; the Go version didn't match the README | Tidied; Go 1.26+ everywhere |
| – | Typos: `minal`, "migratetion", unfinished package comment; misaligned diagram | Fixed |

---

## 3. What the template is now

### Base dependencies
| Package | Purpose |
|---------|---------|
| `github.com/spf13/viper` | Config |
| `github.com/joho/godotenv` | Local `.env` |

**Removed from the base:** `gorm.io/gorm`, `gorm.io/driver/sqlite`, `github.com/mattn/go-sqlite3` (and the interim `glebarez/sqlite` and `pressly/goose`). They're now documented as optional recipes.

### Structure
```
cmd/main.go                        .env → config → slog → signal ctx → BuildApp → app.Run
cmd/di.go                          Composition root (commented 3-step wiring guide)
config/{config,dev,prod}.yaml      Layered config; commented examples for adapter sections
internal/config/                   Loader + schema (App: name, env, shutdowntimeout) + tests
internal/core/domain/              doc.go + errors.go (ErrNotFound, ErrInvalidInput, ErrConflict)
internal/core/port/                doc.go (driving vs driven ports, examples)
internal/core/service/             doc.go (service pattern, example)
internal/adapters/                 doc.go (suggested layout + rules)
internal/platform/lifecycle/       Server interface, App (AddServer/OnStop/Run), HTTPServer + tests
docs/recipes.md                    Tested adapters for frameworks and tools
```

### New: `internal/platform/lifecycle`
This is what makes the template work with any framework. It uses only the standard library.

- **`Server` interface**: `Start(ctx) error` and `Shutdown(ctx) error`. HTTP servers, Fiber, gRPC, workers and queue consumers all fit.
- **`HTTPServer`**: wraps any `http.Handler` (net/http, Gin, Echo, Chi) with safe default timeouts.
- **`App.OnStop`**: registers cleanup for resources (database pools, Redis clients, ...).
- **`App.Run(ctx)`**:
  - starts every server
  - stops everything on `SIGINT`/`SIGTERM` **or** when any server crashes
  - shuts servers down in reverse order, then runs `OnStop` hooks in reverse order (like `defer`), within `app.shutdowntimeout`
  - returns all errors combined with `errors.Join`
- **Tests:** shutdown on signal (checks the order), a server failure stopping the others, no servers registered, hook errors, the shutdown timeout, and a real HTTP round trip.

### Kept from phase 1
- `context.Context`-first port convention (documented in `port/doc.go`)
- Domain errors mapped to responses at the adapter boundary
- `slog` logging (text in dev, JSON otherwise), graceful shutdown
- Config tests (base load, environment merge, env-variable override, unknown environment)
- `Makefile` (`run`, `build`, `test`, `lint`, `fmt`, `tidy`, `docker-build`, `clean`), `.golangci.yml`, GitHub Actions CI
- `Dockerfile`: multi-stage, static binary, `distroless/static:nonroot`. `EXPOSE` is commented out until you add an HTTP server.

---

## 4. Recipes (`docs/recipes.md`)

The recipes are generated from source files that were compiled and linted, so the docs match code that builds.

| Recipe | Library (tested version) | Plugs in via | Run end to end? |
|--------|--------------------------|--------------|-----------------|
| Example feature (Order domain, ports, service) | stdlib | – | ✅ |
| HTTP: net/http | stdlib | `lifecycle.NewHTTPServer` | ✅ |
| HTTP: Gin | gin v1.12.0 | `lifecycle.NewHTTPServer` | ✅ |
| HTTP: Echo | echo v4.16.0 | `lifecycle.NewHTTPServer` | ✅ |
| HTTP: Fiber | fiber v3.6.1 | implements `lifecycle.Server` | ✅ |
| Background worker / consumer | stdlib | implements `lifecycle.Server` | ✅ |
| PostgreSQL | pgx v5.11.0 | `OnStop(pool.Close)` | Compile + lint only |
| Redis cache (`port.Cache`) | go-redis v9.23.0 | `OnStop(rdb.Close)` | Compile + lint only |
| MinIO / S3 (`port.ObjectStorage`) | minio-go v7.4.1 | – (no Close needed) | Compile + lint only |
| Full `di.go` wiring all of the above + config | – | – | Compile + lint |

The recipes also cover gRPC (same pattern as Fiber), GORM, and migrations (goose or golang-migrate).

---

## 5. File changes vs `main`

**Modified:** `.env.example`, `.gitignore`, `README.md`, `cmd/di.go`, `cmd/main.go`, `config/config.yaml`, `config/prod.yaml`, `go.mod`, `go.sum`, `internal/config/config.go`, `internal/config/schema.go`

**Added:**
```
.dockerignore
.github/workflows/ci.yml
.golangci.yml
Dockerfile
Makefile
CHANGES.md
docs/recipes.md
internal/config/config_test.go
internal/core/domain/doc.go
internal/core/domain/errors.go
internal/core/port/doc.go
internal/core/service/doc.go
internal/adapters/doc.go
internal/platform/lifecycle/lifecycle.go
internal/platform/lifecycle/http.go
internal/platform/lifecycle/lifecycle_test.go
```

**Removed:**
```
internal/adapters/repository/sqliter/book_repository_sqlite.go
internal/core/domain/dto/book.go
internal/core/domain/model/book.go
internal/core/port/book_repository.go
internal/core/port/book_service.go
internal/core/service/book_service.go
```

---

## 6. Next steps

1. Run `make docker-build` once to check the Dockerfile.
2. Optionally, try the PostgreSQL, Redis and MinIO recipes against local containers.
3. Review the diff, then commit and open a PR to `main`.
4. Keep real secrets and customer data out of `.env.example`, the YAML files and Git; use environment variables or a secret manager in production.
