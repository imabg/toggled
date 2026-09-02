// Package config loads application configuration from a YAML file via Viper.
package config

import (
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const DefaultPath = "config.yaml"

// Env is the runtime environment. It selects log encoding and can grow
// additional environment-specific behaviour later.
type Env string

const (
	EnvLocal      Env = "local"
	EnvProduction Env = "production"
)

// DatabaseType selects which nested database block is active.
// Add a new constant and a sibling YAML block when introducing another engine.
type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
)

// Config is the root configuration document.
type Config struct {
	Env      Env            `mapstructure:"env"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig is a typed backend selector.
//
//	database:
//	  type: postgres
//	  postgres:
//	    url: postgres://...
//
// A future MySQL backend would add `type: mysql` and a `mysql:` block
// without changing existing keys.
type DatabaseConfig struct {
	Type     DatabaseType   `mapstructure:"type"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

// PostgresConfig holds Postgres connection settings.
type PostgresConfig struct {
	URL string `mapstructure:"url"`
}

// LogConfig holds logger settings.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// Load reads and validates configuration from path.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.ErrorUnused = true
	}); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	var errs []error

	switch c.Env {
	case EnvLocal, EnvProduction:
	default:
		errs = append(errs, fmt.Errorf("config: env must be %q or %q", EnvLocal, EnvProduction))
	}

	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		errs = append(errs, fmt.Errorf("config: http.port must be between 1 and 65535"))
	}

	switch c.Database.Type {
	case DatabaseTypePostgres:
		if c.Database.Postgres.URL == "" {
			errs = append(errs, fmt.Errorf("config: database.postgres.url is required"))
		}
	default:
		errs = append(errs, fmt.Errorf("config: unsupported database.type %q", c.Database.Type))
	}

	if _, err := parseLevel(c.Log.Level); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// PostgresURL returns the active Postgres DSN.
func (c *Config) PostgresURL() string {
	return c.Database.Postgres.URL
}
