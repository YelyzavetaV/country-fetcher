package config

import (
	"os"

	env "github.com/caarlos0/env/v11"
)

type Config struct {
// HTTP config
	HTTPTimeout        string      `env:"HTTP_TIMEOUT" envDefault:"10s"`
// JSON output config
	JSONPrefix         string      `env:"JSON_PREFIX" envDefault:""`
	JSONIndent         string      `env:"JSON_INDENT" envDefault:"  "`
	JSONFilePermission os.FileMode `env:"JSON_FILE_PERMISSION" envDefault:"420"`
	JSONForceOverride  bool        `env:"JSON_FORCE_OVERRIDE" envDefault:"true"`
// Logging
	LogLevel           string      `env:"LOG_LEVEL" envDefault:"INFO"`
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