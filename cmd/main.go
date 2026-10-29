package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// .env is optional: in containers/production, values come from real env vars.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load(os.Getenv("APP_ENV"))
	if err != nil {
		return err
	}

	logger := newLogger(cfg.App.Env)
	slog.SetDefault(logger)

	// ctx is cancelled on Ctrl+C / SIGTERM, which triggers graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := BuildApp(cfg, logger)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}

	logger.Info("application starting", "name", cfg.App.Name, "env", cfg.App.Env)
	return app.Run(ctx)
}

// newLogger returns a human-readable logger for dev and JSON logs elsewhere.
func newLogger(env string) *slog.Logger {
	if env == "dev" {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
