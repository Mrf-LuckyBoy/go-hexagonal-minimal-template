// Package config loads and describes the application configuration.
package config

import "time"

// Config is the root configuration. Add one section per adapter you plug in
// (e.g. HTTP, Redis, MinIO) and a matching block in config/config.yaml.
type Config struct {
	App App `mapstructure:"app"`
}

type App struct {
	Name            string        `mapstructure:"name"`
	Env             string        `mapstructure:"env"`
	ShutdownTimeout time.Duration `mapstructure:"shutdowntimeout"`
}
