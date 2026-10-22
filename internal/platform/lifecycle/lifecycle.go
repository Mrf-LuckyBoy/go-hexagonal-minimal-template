// Package lifecycle runs the application's long-running components and shuts
// them down gracefully.
//
// It is framework-agnostic: anything that can start and stop is a Server
// (net/http, Gin, Echo, Fiber, gRPC, a queue consumer, a cron runner), and
// any resource that must be released (database pool, Redis client, ...) is
// registered with OnStop.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"
)

// Server is a long-running component.
type Server interface {
	// Start runs the component and blocks until it stops. ctx is cancelled
	// when the application begins shutting down. Return nil when the
	// component was stopped on purpose (e.g. by Shutdown).
	Start(ctx context.Context) error

	// Shutdown stops the component gracefully, finishing in-flight work
	// before ctx expires. Components that stop on ctx cancellation alone can
	// simply return nil.
	Shutdown(ctx context.Context) error
}

// StopFunc releases a resource during shutdown.
type StopFunc func(ctx context.Context) error

type namedServer struct {
	name   string
	server Server
}

type stopHook struct {
	name string
	fn   StopFunc
}

// App owns the registered servers and stop hooks.
type App struct {
	logger          *slog.Logger
	shutdownTimeout time.Duration
	servers         []namedServer
	hooks           []stopHook
}

// New creates an App. shutdownTimeout bounds the whole shutdown sequence.
func New(logger *slog.Logger, shutdownTimeout time.Duration) *App {
	return &App{logger: logger, shutdownTimeout: shutdownTimeout}
}

// AddServer registers a long-running component.
func (a *App) AddServer(name string, s Server) {
	a.servers = append(a.servers, namedServer{name: name, server: s})
}

// OnStop registers a cleanup function. Hooks run after all servers have
// stopped, in reverse registration order (like defer), so register a
// resource right after creating it.
func (a *App) OnStop(name string, fn StopFunc) {
	a.hooks = append(a.hooks, stopHook{name: name, fn: fn})
}

// Run starts every server and blocks until ctx is cancelled (e.g. SIGTERM)
// or any server stops on its own. It then shuts everything down and returns
// the combined errors.
func (a *App) Run(ctx context.Context) error {
	if len(a.servers) == 0 {
		a.logger.Warn("no servers registered; register them in cmd/di.go")
		return a.stop(ctx)
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	exited := make(chan error, len(a.servers))
	for _, s := range a.servers {
		wg.Go(func() {
			a.logger.Info("server starting", "server", s.name)
			err := s.server.Start(runCtx)
			if err != nil {
				err = fmt.Errorf("server %q: %w", s.name, err)
			}
			exited <- err
		})
	}

	var runErr error
	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	case runErr = <-exited:
		a.logger.Warn("server stopped unexpectedly; shutting down", "error", runErr)
	}
	cancel()

	return errors.Join(runErr, a.shutdown(&wg))
}

// shutdown stops servers (in reverse order), waits for them, then runs hooks.
func (a *App) shutdown(wg *sync.WaitGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	var errs []error
	for _, s := range slices.Backward(a.servers) {
		if err := s.server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown server %q: %w", s.name, err))
		}
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
		errs = append(errs, fmt.Errorf("servers did not stop within %s: %w", a.shutdownTimeout, ctx.Err()))
	}

	errs = append(errs, a.stop(ctx))
	if err := errors.Join(errs...); err != nil {
		return err
	}
	a.logger.Info("shutdown complete")
	return nil
}

// stop runs the stop hooks in reverse registration order.
func (a *App) stop(ctx context.Context) error {
	var errs []error
	for _, h := range slices.Backward(a.hooks) {
		if err := h.fn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stop %q: %w", h.name, err))
		}
	}
	return errors.Join(errs...)
}
