package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("example file", func(t *testing.T) {
		cfg, err := Load(filepath.Join("..", "..", "config.example.yaml"))
		if err != nil {
			t.Fatalf("Load(config.example.yaml) unexpected error: %v", err)
		}
		if cfg.Env != EnvLocal {
			t.Errorf("Env = %q, want %q", cfg.Env, EnvLocal)
		}
		if cfg.HTTP.Port != 8080 {
			t.Errorf("HTTP.Port = %d, want 8080", cfg.HTTP.Port)
		}
		if cfg.Database.Type != DatabaseTypePostgres {
			t.Errorf("Database.Type = %q, want %q", cfg.Database.Type, DatabaseTypePostgres)
		}
		if cfg.PostgresURL() == "" {
			t.Error("PostgresURL() is empty")
		}
		if cfg.Log.Level != "info" {
			t.Errorf("Log.Level = %q, want info", cfg.Log.Level)
		}
	})

	t.Run("valid yaml", func(t *testing.T) {
		path := writeConfig(t, `
env: local
http:
  port: 9090
database:
  type: postgres
  postgres:
    url: postgres://toggled:toggled@localhost:5432/toggled?sslmode=disable
log:
  level: debug
`)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() unexpected error: %v", err)
		}
		if cfg.Env != EnvLocal {
			t.Errorf("Env = %q, want %q", cfg.Env, EnvLocal)
		}
		if cfg.HTTP.Port != 9090 {
			t.Errorf("HTTP.Port = %d, want 9090", cfg.HTTP.Port)
		}
		if cfg.Database.Type != DatabaseTypePostgres {
			t.Errorf("Database.Type = %q, want %q", cfg.Database.Type, DatabaseTypePostgres)
		}
		wantURL := "postgres://toggled:toggled@localhost:5432/toggled?sslmode=disable"
		if cfg.PostgresURL() != wantURL {
			t.Errorf("PostgresURL() = %q, want %q", cfg.PostgresURL(), wantURL)
		}
		if cfg.Log.Level != "debug" {
			t.Errorf("Log.Level = %q, want debug", cfg.Log.Level)
		}
	})

	t.Run("omitted fields are not defaulted", func(t *testing.T) {
		path := writeConfig(t, `
database:
  type: postgres
  postgres:
    url: postgres://localhost/toggled
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want validation error for omitted fields")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
		if err == nil {
			t.Fatal("Load() error = nil, want error")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		path := writeConfig(t, "http: [\n")
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want parse error")
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		path := writeConfig(t, `
database:
  postgres:
    url: postgres://localhost/toggled
mystery: true
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want unknown field error")
		}
	})

	t.Run("missing postgres url", func(t *testing.T) {
		path := writeConfig(t, `
database:
  type: postgres
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want validation error")
		}
	})

	t.Run("unsupported database type", func(t *testing.T) {
		path := writeConfig(t, `
database:
  type: mysql
  postgres:
    url: postgres://localhost/toggled
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want unsupported type error")
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		path := writeConfig(t, `
http:
  port: 70000
database:
  postgres:
    url: postgres://localhost/toggled
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want port validation error")
		}
	})

	t.Run("invalid env", func(t *testing.T) {
		path := writeConfig(t, `
env: staging
database:
  postgres:
    url: postgres://localhost/toggled
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want env validation error")
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		path := writeConfig(t, `
database:
  postgres:
    url: postgres://localhost/toggled
log:
  level: verbose
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want log level error")
		}
	})
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
