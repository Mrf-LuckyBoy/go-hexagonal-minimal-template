package main

import (
	"log/slog"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/config"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/platform/lifecycle"
)

// BuildApp wires adapters -> services -> driving adapters and registers them
// with the application lifecycle. This is the only place that knows about
// concrete implementations. See the README "Recipes" section for examples.
func BuildApp(cfg *config.Config, logger *slog.Logger) (*lifecycle.App, error) {
	app := lifecycle.New(logger, cfg.App.ShutdownTimeout)

	// 1. Driven adapters: databases, caches, object storage, external APIs.
	//    Register cleanup right after creating each one:
	//
	//	rdb, err := redisadapter.NewClient(cfg.Redis)
	//	if err != nil { return nil, err }
	//	app.OnStop("redis", func(context.Context) error { return rdb.Close() })
	//	cache := redisadapter.NewCache(rdb)

	// 2. Services: business logic built from driven ports.
	//
	//	orderService := service.NewOrderService(orderRepository, cache)

	// 3. Driving adapters: HTTP/gRPC servers, consumers, schedulers.
	//
	//	router := ginhandler.NewRouter(orderService)
	//	app.AddServer("http", lifecycle.NewHTTPServer(":8080", router, 30*time.Second))

	return app, nil
}
