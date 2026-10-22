package lifecycle_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/platform/lifecycle"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

// fakeServer blocks in Start until Shutdown is called or ctx is cancelled,
// unless startErr is set, in which case Start fails immediately.
type fakeServer struct {
	startErr error
	stopped  chan struct{}
	once     sync.Once
	log      *recorder
	name     string
}

func newFakeServer(name string, log *recorder) *fakeServer {
	return &fakeServer{name: name, log: log, stopped: make(chan struct{})}
}

func (f *fakeServer) Start(ctx context.Context) error {
	if f.startErr != nil {
		return f.startErr
	}
	select {
	case <-f.stopped:
	case <-ctx.Done():
	}
	return nil
}

func (f *fakeServer) Shutdown(context.Context) error {
	f.log.add("shutdown " + f.name)
	f.once.Do(func() { close(f.stopped) })
	return nil
}

// recorder collects events from concurrent goroutines.
type recorder struct {
	mu     sync.Mutex
	events []string
}

func (r *recorder) add(e string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

func (r *recorder) get() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.events)
}

func (r *recorder) hook(name string) lifecycle.StopFunc {
	return func(context.Context) error {
		r.add("stop " + name)
		return nil
	}
}

func TestRun_ShutdownOnContextCancel(t *testing.T) {
	rec := &recorder{}
	app := lifecycle.New(discard, time.Second)
	app.OnStop("db", rec.hook("db"))
	app.OnStop("cache", rec.hook("cache"))
	app.AddServer("http", newFakeServer("http", rec))
	app.AddServer("worker", newFakeServer("worker", rec))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- app.Run(ctx) }()
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	// Servers shut down in reverse order, then hooks run in reverse order.
	want := []string{"shutdown worker", "shutdown http", "stop cache", "stop db"}
	if got := rec.get(); !slices.Equal(got, want) {
		t.Errorf("events = %v, want %v", got, want)
	}
}

func TestRun_ServerFailureStopsEverything(t *testing.T) {
	rec := &recorder{}
	boom := errors.New("port in use")
	failing := newFakeServer("http", rec)
	failing.startErr = boom

	app := lifecycle.New(discard, time.Second)
	app.OnStop("db", rec.hook("db"))
	app.AddServer("worker", newFakeServer("worker", rec))
	app.AddServer("http", failing)

	err := app.Run(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("Run() error = %v, want %v", err, boom)
	}
	if got := rec.get(); !slices.Contains(got, "shutdown worker") || !slices.Contains(got, "stop db") {
		t.Errorf("events = %v, want worker shut down and db stopped", got)
	}
}

func TestRun_NoServers(t *testing.T) {
	rec := &recorder{}
	app := lifecycle.New(discard, time.Second)
	app.OnStop("db", rec.hook("db"))

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := rec.get(); !slices.Equal(got, []string{"stop db"}) {
		t.Errorf("events = %v, want hooks to still run", got)
	}
}

func TestRun_HookErrorsAreReturned(t *testing.T) {
	hookErr := errors.New("close failed")
	app := lifecycle.New(discard, time.Second)
	app.OnStop("db", func(context.Context) error { return hookErr })
	app.AddServer("http", newFakeServer("http", &recorder{}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := app.Run(ctx); !errors.Is(err, hookErr) {
		t.Fatalf("Run() error = %v, want %v", err, hookErr)
	}
}

// stuckServer ignores Shutdown and ctx, to exercise the shutdown timeout.
type stuckServer struct{ release chan struct{} }

func (s stuckServer) Start(context.Context) error    { <-s.release; return nil }
func (s stuckServer) Shutdown(context.Context) error { return nil }

func TestRun_ShutdownTimeout(t *testing.T) {
	stuck := stuckServer{release: make(chan struct{})}
	defer close(stuck.release)

	app := lifecycle.New(discard, 50*time.Millisecond)
	app.AddServer("stuck", stuck)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := app.Run(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestHTTPServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "pong") })

	srv := &http.Server{Addr: "127.0.0.1:18089", Handler: mux, ReadHeaderTimeout: time.Second}
	app := lifecycle.New(discard, time.Second)
	app.AddServer("http", lifecycle.NewHTTPServerFrom(srv))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- app.Run(ctx) }()

	var resp *http.Response
	var err error
	for range 50 {
		if resp, err = http.Get("http://" + srv.Addr + "/ping"); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("GET /ping error = %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != "pong" {
		t.Errorf("body = %q, want %q", body, "pong")
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
