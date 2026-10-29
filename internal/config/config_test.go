package config_test

import (
	"testing"
	"time"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/config"
)

func TestLoad(t *testing.T) {
	t.Chdir("../..") // Load reads ./config relative to the repo root

	t.Run("base config", func(t *testing.T) {
		cfg, err := config.Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.App.Name != "myapp" || cfg.App.ShutdownTimeout != 10*time.Second {
			t.Errorf("Load() = %+v", cfg)
		}
	})

	t.Run("env file is merged", func(t *testing.T) {
		cfg, err := config.Load("prod")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.App.Env != "prod" {
			t.Errorf("Load(prod) app.env = %q, want %q", cfg.App.Env, "prod")
		}
		if cfg.App.Name != "myapp" {
			t.Errorf("Load(prod) lost base value app.name = %q", cfg.App.Name)
		}
	})

	t.Run("environment variables override files", func(t *testing.T) {
		t.Setenv("APP_NAME", "from-env")
		t.Setenv("APP_SHUTDOWNTIMEOUT", "3s")
		cfg, err := config.Load("dev")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.App.Name != "from-env" || cfg.App.ShutdownTimeout != 3*time.Second {
			t.Errorf("Load() = %+v", cfg.App)
		}
	})

	t.Run("unknown env file fails", func(t *testing.T) {
		if _, err := config.Load("nope"); err == nil {
			t.Error("Load(nope) error = nil, want error")
		}
	})
}
