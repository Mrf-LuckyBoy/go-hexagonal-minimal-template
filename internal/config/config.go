package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Load(env string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read base config error: %w", err)
	}

	fmt.Println("Using config:", v.ConfigFileUsed())

	if env != "" {
		v.SetConfigName(env)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge env config error: %w", err)
		}
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
