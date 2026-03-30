// Package config use for config value of project
package config

type Config struct {
	App      App      `mapstructure:"app"`
	Database Database `mapstructure:"database"`
	Server   Server   `mapstructure:"server"`
}

type App struct {
	Name       string `mapstructure:"name"`
	Port       int    `mapstructure:"port"`
	Env        string `mapstructure:"env"`
	EncryptKey string `mapstructure:"encryptkey"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type Server struct {
	Timeout string `mapstructure:"timeout"`
}
