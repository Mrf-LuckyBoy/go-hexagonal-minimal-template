package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load reads config/config.yaml, merges config/<env>.yaml on top when env is
// not empty, then applies environment variable overrides
// (e.g. database.password -> DATABASE_PASSWORD).
//
// Note: environment overrides only apply to keys that exist in the YAML files,
// so every secret needs a placeholder key in config/config.yaml.
func Load(env string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read base config: %w", err)
	}

	if env != "" {
		v.SetConfigName(env)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge %q config: %w", env, err)
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &cfg, nil
}
