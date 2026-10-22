# go-hexagonal-minimal-template

A **bare** Go service skeleton following **Hexagonal Architecture** (Ports & Adapters). It doesn't pick a framework for you. Plug in **any** HTTP framework (net/http, Gin, Echo, Fiber, gRPC) and **any** infrastructure (PostgreSQL, Redis, MinIO, Kafka, ...) as adapters, without touching your business logic.

The base has only two dependencies: [Viper](https://github.com/spf13/viper) (config) and [godotenv](https://github.com/joho/godotenv) (local `.env`). Everything else is opt-in.

---

## Features

- **Hexagonal skeleton**: `core/` (domain, ports, services) is kept separate from `adapters/`, with doc files that explain what goes where
- **Lifecycle manager** (`internal/platform/lifecycle`): any framework, worker or consumer plugs in as a `Server`; any resource registers cleanup with `OnStop`
- **Graceful shutdown**: on `SIGINT`/`SIGTERM`, servers stop in reverse order, then resources close, all within a configurable timeout. If one server crashes, everything else shuts down cleanly.
- **Layered config**: base YAML → environment YAML → environment variables
- **Structured logging** with `log/slog` (text in dev, JSON otherwise)
- **Shared domain errors** (`ErrNotFound`, `ErrInvalidInput`, `ErrConflict`) for consistent error mapping across adapters
- **Tooling**: `Makefile`, `golangci-lint` config, `Dockerfile` (distroless, non-root) and GitHub Actions CI
- **[Recipes](docs/recipes.md)**: tested, copy-paste adapters for **net/http, Gin, Echo, Fiber, workers, PostgreSQL, Redis and MinIO**

---

## 📁 Project Structure

```
.
├── cmd/
│   ├── main.go                   # Entry point: .env, config, logger, signals, run
│   └── di.go                     # Composition root: wire adapters → services → servers
│
├── config/
│   ├── config.yaml               # Base configuration
│   ├── dev.yaml                  # Overrides when APP_ENV=dev
│   └── prod.yaml                 # Overrides when APP_ENV=prod
│
├── internal/
│   ├── config/                   # Config loader (Viper) + schema
│   │
│   ├── core/                     # ← The Hexagon: business logic, standard library only
│   │   ├── domain/               # Entities, value objects, domain errors
│   │   ├── port/                 # Interfaces: driving (use cases) + driven (dependencies)
│   │   └── service/              # Use-case implementations
│   │
│   ├── adapters/                 # ← Adapters: frameworks & infrastructure (empty, add your own)
│   │
│   └── platform/
│       └── lifecycle/            # Start/stop servers & resources, graceful shutdown
│
├── docs/recipes.md               # Gin / Echo / Fiber / Redis / MinIO / PostgreSQL / worker
├── Dockerfile
├── Makefile
└── .golangci.yml
```

`core/` and `adapters/` start empty. Each folder has a `doc.go` that explains what belongs there, with examples.

---

## 🏛️ Architecture

```
       Driving adapters                  Core                     Driven adapters
   ┌────────────────────┐    ┌──────────────────────────┐    ┌────────────────────┐
   │ net/http · Gin     │    │  port (driving)          │    │ PostgreSQL · GORM  │
   │ Echo · Fiber       │───►│      ▲                   │    │ Redis              │
   │ gRPC · consumers   │    │  service ─► port (driven)│◄───│ MinIO / S3         │
   │ cron · CLI         │    │      │                   │    │ external APIs      │
   └────────────────────┘    │  domain (+ errors)       │    └────────────────────┘
            ▲                └──────────────────────────┘              ▲
            │                                                         │
            └────────────── cmd/di.go wires everything ────────────────┘
                     platform/lifecycle starts & stops it all
```

**Rules**
1. `core/` imports only the standard library and other `core/` packages. No frameworks, no drivers, no struct tags.
2. The core **owns** the ports. Adapters implement driven ports, and driving adapters call driving ports.
3. Adapters never import each other. Only `cmd/di.go` knows about concrete types.
4. Adapters translate infrastructure errors into domain errors at the boundary (e.g. `redis.Nil` → `domain.ErrNotFound`). Driving adapters map domain errors to responses (e.g. `ErrNotFound` → 404).
5. Every port method takes `context.Context` first.

---

## ♻️ Lifecycle: plugging in anything

```go
// internal/platform/lifecycle
type Server interface {
    Start(ctx context.Context) error    // block until stopped; ctx is cancelled on shutdown
    Shutdown(ctx context.Context) error // stop gracefully before ctx expires
}

app := lifecycle.New(logger, cfg.App.ShutdownTimeout)
app.OnStop("redis", func(context.Context) error { return rdb.Close() }) // resources
app.AddServer("http", lifecycle.NewHTTPServer(":8080", router, 30*time.Second))
app.AddServer("consumer", myConsumer)                               // anything with Start/Shutdown
return app.Run(ctx)
```

| You have | How it plugs in |
|---|---|
| Any `http.Handler`: net/http, **Gin**, **Echo**, Chi | `lifecycle.NewHTTPServer(addr, handler, timeout)` |
| A framework not based on `net/http`: **Fiber**, gRPC | Implement `Server` (2 methods). See [recipes](docs/recipes.md#http-fiber). |
| A worker, cron job or queue consumer (Kafka, RabbitMQ, NATS) | Implement `Server`, looping until `ctx` is done |
| A resource to close: DB pool, **Redis**, file handles | `app.OnStop(name, fn)` |

**Shutdown order:** servers stop in reverse registration order, then `OnStop` hooks run in reverse order (like `defer`), all within `app.shutdowntimeout`.

---

## 🚀 Getting Started

**Prerequisites:** Go 1.26+, and optionally [golangci-lint](https://golangci-lint.run/) v2 and Docker.

```bash
git clone https://github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template.git my-service
cd my-service
cp .env.example .env   # optional
make run
```

The bare skeleton starts, warns `no servers registered`, and exits. That's expected. Add your first adapter in `cmd/di.go` and it will keep running until it receives a shutdown signal.

```
make run           # run locally
make test          # tests with -race and coverage
make lint          # golangci-lint
make fmt           # gofmt + goimports
make build         # static binary in bin/
make docker-build  # build the Docker image
```

---

## 🔌 Adding a Feature

1. **Domain:** `internal/core/domain/order.go`, plain structs.
2. **Ports:** `internal/core/port/order.go`, with a driving `OrderService` and a driven `OrderRepository`.
3. **Service:** `internal/core/service/order_service.go`, plus a unit test using an in-memory fake of the driven port.
4. **Adapters:** e.g. `internal/adapters/repository/postgres/` and `internal/adapters/handler/gin/`.
5. **Config:** add sections to `internal/config/schema.go` and `config/config.yaml`.
6. **Wire** it in `cmd/di.go`.

[docs/recipes.md](docs/recipes.md) walks through all of this with complete code, including a full `di.go` that uses PostgreSQL, Redis, MinIO, Gin and a background worker.

---

## ⚙️ Configuration

Layers, each overriding the previous one:

1. `config/config.yaml`: base defaults
2. `config/<APP_ENV>.yaml`: e.g. `APP_ENV=prod` loads `config/prod.yaml`
3. **Environment variables**: `UPPER_SNAKE_CASE` of the key, e.g. `redis.password` → `REDIS_PASSWORD`

A `.env` file is loaded if one exists. It's for local development only: **never commit it**. In production, use real environment variables or a secret manager.

> ⚠️ An environment variable only overrides a key that **already exists in the YAML**. Add an empty placeholder for every secret in `config/config.yaml`.

---

## 🛠️ Using This Template

### Rename the Go module

```bash
OLD=github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template
NEW=github.com/YourUsername/your-project-name

go mod edit -module "$NEW"
# Linux (macOS: use sed -i '' instead)
grep -rl "$OLD" --include='*.go' --include='*.md' --include='.golangci.yml' . | xargs sed -i "s|$OLD|$NEW|g"

grep -r "$OLD" . --exclude-dir=.git   # should print nothing
go mod tidy && go build ./... && go test ./...
```

Then update the `LICENSE` holder, `app.name` in `config/config.yaml`, and this README.

---

## 📦 Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/viper` | Config management |
| `github.com/joho/godotenv` | `.env` loading for local dev |

---

## 📝 License

MIT. See [LICENSE](./LICENSE).
