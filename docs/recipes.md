# Recipes

Copy-paste adapters for common frameworks and tools. Each recipe follows the same steps:

1. **Port**: define the interface the core needs, in `internal/core/port/`.
2. **Adapter**: implement it in `internal/adapters/<role>/<tool>/`.
3. **Config**: add a section to `internal/config/schema.go` and `config/config.yaml`. Leave secrets empty and set them through environment variables.
4. **Wire**: create the adapter in `cmd/di.go`, register cleanup with `app.OnStop`, and register long-running components with `app.AddServer`.
5. `go get <library>` and `go mod tidy`.

> Every snippet below was compiled and linted against the library version shown. The HTTP framework recipes and the worker were also run end to end: they served requests, returned 404 for a missing order, and shut down gracefully. **Note:** the template's module path is used in imports. If you renamed your module, replace it.

**Contents**
- [The example feature (Order)](#the-example-feature-order)
- [HTTP: net/http](#http-nethttp-standard-library)
- [HTTP: Gin](#http-gin)
- [HTTP: Echo](#http-echo)
- [HTTP: Fiber](#http-fiber)
- [Background worker / queue consumer](#background-worker--queue-consumer)
- [PostgreSQL (pgx)](#postgresql-pgx)
- [Redis (cache)](#redis-cache)
- [MinIO / S3 (object storage)](#minio--s3-object-storage)
- [Wiring it all together](#wiring-it-all-together)
- [Config for all recipes](#config-for-all-recipes)

---

## The example feature (Order)

The recipes share one small feature so the adapters have something to call.

`internal/core/domain/order.go`
```go
package domain

import "time"

type Order struct {
	ID         string
	CustomerID string
	Total      int64 // minor units (satang/cents)
	CreatedAt  time.Time
}
```

`internal/core/port/order.go`
```go
package port

import (
	"context"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
)

// Driving port
type OrderService interface {
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
}

// Driven port
type OrderRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Order, error) // domain.ErrNotFound if missing
}
```

`internal/core/service/order_service.go`
```go
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

type orderService struct {
	repo port.OrderRepository
}

func NewOrderService(repo port.OrderRepository) port.OrderService {
	return &orderService{repo: repo}
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order %q: %w", id, err)
	}
	return order, nil
}
```

---

## HTTP: net/http (standard library)

No dependency. Plug it in with `lifecycle.NewHTTPServer`.

`internal/adapters/handler/nethttp/router.go`
```go
// Package nethttp is an HTTP driving adapter using only the standard library.
package nethttp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

func NewRouter(orders port.OrderService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		order, err := orders.GetOrder(r.Context(), r.PathValue("id"))
		if err != nil {
			writeJSON(w, statusFor(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, order) // use a response DTO in real code
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// statusFor maps domain errors to HTTP status codes.
func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
```

```go
// cmd/di.go
app.AddServer("http", lifecycle.NewHTTPServer(":8080", nethttp.NewRouter(orderService), 30*time.Second))
```

---

## HTTP: Gin

`go get github.com/gin-gonic/gin` (checked with v1.12.0). `*gin.Engine` implements `http.Handler`, so the lifecycle code doesn't change.

`internal/adapters/handler/gin/router.go`
```go
// Package ginhandler is an HTTP driving adapter using Gin.
package ginhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

// NewRouter returns a *gin.Engine, which implements http.Handler,
// so it plugs into lifecycle.NewHTTPServer unchanged.
func NewRouter(orders port.OrderService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/orders/:id", func(c *gin.Context) {
		order, err := orders.GetOrder(c.Request.Context(), c.Param("id"))
		if err != nil {
			c.JSON(statusFor(err), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, order)
	})
	return r
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
```

```go
// cmd/di.go
app.AddServer("http", lifecycle.NewHTTPServer(":8080", ginhandler.NewRouter(orderService), 30*time.Second))
```

> Set `GIN_MODE=release` in production.

---

## HTTP: Echo

`go get github.com/labstack/echo/v4` (checked with v4.16.0). `*echo.Echo` implements `http.Handler`.

`internal/adapters/handler/echo/router.go`
```go
// Package echohandler is an HTTP driving adapter using Echo.
package echohandler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

// NewRouter returns an *echo.Echo, which implements http.Handler,
// so it plugs into lifecycle.NewHTTPServer unchanged.
func NewRouter(orders port.OrderService) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())

	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/orders/:id", func(c echo.Context) error {
		order, err := orders.GetOrder(c.Request().Context(), c.Param("id"))
		if err != nil {
			return c.JSON(statusFor(err), map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, order)
	})
	return e
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
```

```go
// cmd/di.go
app.AddServer("http", lifecycle.NewHTTPServer(":8080", echohandler.NewRouter(orderService), 30*time.Second))
```

---

## HTTP: Fiber

`go get github.com/gofiber/fiber/v3` (checked with v3.6.1). Fiber is built on fasthttp, not `net/http`, so the adapter implements `lifecycle.Server` itself. Any framework that isn't based on `net/http` (gRPC, for example) works the same way.

`internal/adapters/handler/fiber/server.go`
```go
// Package fiberhandler is an HTTP driving adapter using Fiber. Fiber is not
// built on net/http, so it implements lifecycle.Server itself.
package fiberhandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/platform/lifecycle"
)

var _ lifecycle.Server = (*Server)(nil)

type Server struct {
	app  *fiber.App
	addr string
}

func NewServer(addr string, orders port.OrderService) *Server {
	app := fiber.New()

	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/orders/:id", func(c fiber.Ctx) error {
		order, err := orders.GetOrder(c.Context(), c.Params("id"))
		if err != nil {
			return c.Status(statusFor(err)).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(order)
	})

	return &Server{app: app, addr: addr}
}

func (s *Server) Start(context.Context) error {
	return s.app.Listen(s.addr, fiber.ListenConfig{DisableStartupMessage: true})
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
```

```go
// cmd/di.go
app.AddServer("http", fiberhandler.NewServer(":8080", orderService))
```

> **gRPC** follows the same pattern: `Start` calls `grpcServer.Serve(listener)` and `Shutdown` calls `grpcServer.GracefulStop()`.

---

## Background worker / queue consumer

Anything long-running is a `lifecycle.Server`. This example runs a job on a timer; a Kafka, RabbitMQ or NATS consumer has the same shape: loop until `ctx` is cancelled.

`internal/adapters/worker/ticker.go`
```go
// Package worker shows a background job as a lifecycle.Server
// (the same shape fits queue consumers, e.g. Kafka/RabbitMQ/NATS).
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/platform/lifecycle"
)

var _ lifecycle.Server = (*Ticker)(nil)

// Ticker calls job every interval until shutdown.
type Ticker struct {
	interval time.Duration
	job      func(ctx context.Context) error
	logger   *slog.Logger
}

func NewTicker(interval time.Duration, job func(ctx context.Context) error, logger *slog.Logger) *Ticker {
	return &Ticker{interval: interval, job: job, logger: logger}
}

func (t *Ticker) Start(ctx context.Context) error {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := t.job(ctx); err != nil {
				t.logger.Error("job failed", "error", err)
			}
		}
	}
}

// Shutdown is a no-op: Start returns when ctx is cancelled.
func (t *Ticker) Shutdown(context.Context) error { return nil }
```

```go
// cmd/di.go
app.AddServer("cleanup-job", worker.NewTicker(time.Minute, cleanupService.Run, logger))
```

---

## PostgreSQL (pgx)

`go get github.com/jackc/pgx/v5` (checked with v5.11.0). The adapter translates `pgx.ErrNoRows` into `domain.ErrNotFound`.

`internal/adapters/repository/postgres/order_repository.go`
```go
// Package postgres implements repository ports with PostgreSQL (pgx).
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

// NewPool connects and verifies the connection.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

var _ port.OrderRepository = (*OrderRepository)(nil)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var o domain.Order
	err := r.pool.QueryRow(ctx,
		"SELECT id, customer_id, total, created_at FROM orders WHERE id = $1", id,
	).Scan(&o.ID, &o.CustomerID, &o.Total, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}
```

```go
// cmd/di.go
pool, err := postgres.NewPool(ctx, cfg.Postgres.DSN)
if err != nil {
	return nil, err
}
app.OnStop("postgres", func(context.Context) error { pool.Close(); return nil })
orderRepository := postgres.NewOrderRepository(pool)
```

> **Migrations:** use [goose](https://github.com/pressly/goose) or [golang-migrate](https://github.com/golang-migrate/migrate). Embed the `.sql` files with `//go:embed` and run them in `BuildApp` before creating repositories, or as a separate step in your deploy pipeline. **GORM** works the same way: open `*gorm.DB` in `di.go`, register `sqlDB.Close` with `OnStop`, and keep GORM-tagged structs inside the adapter.

---

## Redis (cache)

`go get github.com/redis/go-redis/v9` (checked with v9.23.0).

`internal/core/port/cache.go`
```go
package port

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error) // domain.ErrNotFound on miss
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
```

`internal/adapters/cache/redis/cache.go`
```go
// Package rediscache implements port.Cache with Redis (go-redis).
package rediscache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

// NewClient connects and verifies the connection.
func NewClient(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return rdb, nil
}

var _ port.Cache = (*Cache)(nil)

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, domain.ErrNotFound
	}
	return val, err
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}
```

```go
// cmd/di.go
rdb, err := rediscache.NewClient(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
if err != nil {
	return nil, err
}
app.OnStop("redis", func(context.Context) error { return rdb.Close() })
cache := rediscache.NewCache(rdb) // pass to services as port.Cache
```

---

## MinIO / S3 (object storage)

`go get github.com/minio/minio-go/v7` (checked with v7.4.1). This works with MinIO and any S3-compatible store.

`internal/core/port/object_storage.go`
```go
package port

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error) // domain.ErrNotFound if missing
	Delete(ctx context.Context, key string) error
}
```

`internal/adapters/storage/minio/storage.go`
```go
// Package miniostorage implements port.ObjectStorage with MinIO / any S3-compatible store.
package miniostorage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/port"
)

// NewClient creates a client and makes sure the bucket exists.
func NewClient(ctx context.Context, endpoint, accessKey, secretKey, bucket string, useSSL bool) (*minio.Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket %q: %w", bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket %q: %w", bucket, err)
		}
	}
	return client, nil
}

var _ port.ObjectStorage = (*Storage)(nil)

type Storage struct {
	client *minio.Client
	bucket string
}

func NewStorage(client *minio.Client, bucket string) *Storage {
	return &Storage{client: client, bucket: bucket}
}

func (s *Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// GetObject is lazy: Stat surfaces "not found" before the caller reads.
	if _, err := obj.Stat(); err != nil {
		_ = obj.Close()
		if minio.ToErrorResponse(err).Code == minio.NoSuchKey {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return obj, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
```

```go
// cmd/di.go (the MinIO client is HTTP-based and needs no Close)
mc, err := miniostorage.NewClient(ctx, cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, cfg.MinIO.Bucket, cfg.MinIO.UseSSL)
if err != nil {
	return nil, err
}
storage := miniostorage.NewStorage(mc, cfg.MinIO.Bucket) // pass to services as port.ObjectStorage
```

---

## Wiring it all together

A complete `cmd/di.go` using PostgreSQL, Redis, MinIO, Gin and a worker:

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	rediscache "github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/adapters/cache/redis"
	ginhandler "github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/adapters/handler/gin"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/adapters/repository/postgres"
	miniostorage "github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/adapters/storage/minio"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/adapters/worker"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/config"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/service"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/platform/lifecycle"
)

func BuildApp(cfg *config.Config, logger *slog.Logger) (*lifecycle.App, error) {
	app := lifecycle.New(logger, cfg.App.ShutdownTimeout)
	ctx := context.Background()

	// 1. Driven adapters
	pool, err := postgres.NewPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		return nil, err
	}
	app.OnStop("postgres", func(context.Context) error { pool.Close(); return nil })
	orderRepository := postgres.NewOrderRepository(pool)

	rdb, err := rediscache.NewClient(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return nil, err
	}
	app.OnStop("redis", func(context.Context) error { return rdb.Close() })
	cache := rediscache.NewCache(rdb)
	_ = cache // pass to services that need it

	mc, err := miniostorage.NewClient(ctx, cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, cfg.MinIO.Bucket, cfg.MinIO.UseSSL)
	if err != nil {
		return nil, err
	}
	storage := miniostorage.NewStorage(mc, cfg.MinIO.Bucket)
	_ = storage

	// 2. Services
	orderService := service.NewOrderService(orderRepository)

	// 3. Driving adapters
	router := ginhandler.NewRouter(orderService)
	app.AddServer("http", lifecycle.NewHTTPServer(fmt.Sprintf(":%d", cfg.HTTP.Port), router, cfg.HTTP.Timeout))
	app.AddServer("cleanup-job", worker.NewTicker(time.Minute, func(context.Context) error { return nil }, logger))

	return app, nil
}
```

> If `BuildApp` fails partway through (say Redis is down after PostgreSQL connected), `main` exits and the operating system closes any connections already opened. That's the usual approach for startup failures.

---

## Config for all recipes

`internal/config/schema.go`
```go
// Package config loads and describes the application configuration.
package config

import "time"

type Config struct {
	App      App      `mapstructure:"app"`
	HTTP     HTTP     `mapstructure:"http"`
	Postgres Postgres `mapstructure:"postgres"`
	Redis    Redis    `mapstructure:"redis"`
	MinIO    MinIO    `mapstructure:"minio"`
}

type App struct {
	Name            string        `mapstructure:"name"`
	Env             string        `mapstructure:"env"`
	ShutdownTimeout time.Duration `mapstructure:"shutdowntimeout"`
}

type HTTP struct {
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type Postgres struct {
	DSN string `mapstructure:"dsn"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MinIO struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accesskey"`
	SecretKey string `mapstructure:"secretkey"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"usessl"`
}
```

`config/config.yaml`
```yaml
app:
  name: myapp
  env: dev
  shutdowntimeout: 10s

http:
  port: 8080
  timeout: 30s

postgres:
  dsn: ""          # set via POSTGRES_DSN

redis:
  addr: localhost:6379
  password: ""     # set via REDIS_PASSWORD
  db: 0

minio:
  endpoint: localhost:9000
  accesskey: ""    # set via MINIO_ACCESSKEY
  secretkey: ""    # set via MINIO_SECRETKEY
  bucket: uploads
  usessl: false
```

`.env` (local only, never commit)
```bash
POSTGRES_DSN=postgres://user:change-me@localhost:5432/myapp?sslmode=disable
REDIS_PASSWORD=change-me
MINIO_ACCESSKEY=change-me
MINIO_SECRETKEY=change-me
```
