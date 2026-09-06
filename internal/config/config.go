// Package config loads application configuration from a YAML file.
package config

import (
	"errors"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

const DefaultPath = "config.yaml"

// DatabaseURLEnv overrides database.url when set, so the DSN can come from
// the environment (e.g. containers) instead of the config file.
const DatabaseURLEnv = "DATABASE_URL"

// Env is the runtime environment. It selects log encoding and can grow
// additional environment-specific behaviour later.
type Env string

const (
	EnvLocal      Env = "local"
	EnvProduction Env = "production"
)

// Config is the root configuration document. The section structs are
// embedded, so their fields are promoted: cfg.Port, cfg.URL, cfg.Level.
type Config struct {
	Env            Env `yaml:"env"`
	HTTPConfig     `yaml:",inline"`
	DatabaseConfig `yaml:",inline"`
	LogConfig      `yaml:",inline"`
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Port int `yaml:"port"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	URL string `yaml:"url"`
}

// LogConfig holds logger settings.
type LogConfig struct {
	Level string `yaml:"level"`
}

// Load reads and validates configuration from path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	if url := os.Getenv(DatabaseURLEnv); url != "" {
		cfg.URL = url
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

	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("config: port must be between 1 and 65535"))
	}

	if c.URL == "" {
		errs = append(errs, fmt.Errorf("config: url is required"))
	}

	if _, err := parseLevel(c.Level); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
