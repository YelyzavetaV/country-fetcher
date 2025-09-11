package config

import (
	env "github.com/caarlos0/env/v11"
)

type Config struct {
	JSONPrefix string `env:"JSON_PREFIX" envDefault:""`
	JSONIndent string `env:"JSON_INDENT" envDefault:"  "`
}

func NewConfig() *Config {
	var cfg Config
	if err := env.ParseWithOptions(&cfg, env.Options{
		Prefix: "FETCHER_",
	}); err != nil {
		panic(err)
	}
	return &cfg
}