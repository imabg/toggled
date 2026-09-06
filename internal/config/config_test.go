package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("example file", func(t *testing.T) {
		t.Parallel()

		cfg, err := Load(filepath.Join("..", "..", "config.example.yaml"))
		if err != nil {
			t.Fatalf("Load(config.example.yaml) unexpected error: %v", err)
		}
		if cfg.Env != EnvLocal {
			t.Errorf("Env = %q, want %q", cfg.Env, EnvLocal)
		}
		if cfg.Port != 8080 {
			t.Errorf("Port = %d, want 8080", cfg.Port)
		}
		if cfg.URL == "" {
			t.Error("URL is empty")
		}
		if cfg.Level != "info" {
			t.Errorf("Level = %q, want info", cfg.Level)
		}
	})

	t.Run("valid yaml", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, `
env: local
port: 9090
url: postgres://toggled:toggled@localhost:5432/toggled?sslmode=disable
level: debug
`)
		cfg, err := Load(path)
		if err != nil {
			t.Fatalf("Load() unexpected error: %v", err)
		}
		if cfg.Env != EnvLocal {
			t.Errorf("Env = %q, want %q", cfg.Env, EnvLocal)
		}
		if cfg.Port != 9090 {
			t.Errorf("Port = %d, want 9090", cfg.Port)
		}
		wantURL := "postgres://toggled:toggled@localhost:5432/toggled?sslmode=disable"
		if cfg.URL != wantURL {
			t.Errorf("URL = %q, want %q", cfg.URL, wantURL)
		}
		if cfg.Level != "debug" {
			t.Errorf("Level = %q, want debug", cfg.Level)
		}
	})

	t.Run("omitted fields are not defaulted", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, `
url: postgres://localhost/toggled
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want validation error for omitted fields")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
		if err == nil {
			t.Fatal("Load() error = nil, want error")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, "port: [\n")
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want parse error")
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, `
url: postgres://localhost/toggled
mystery: true
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want unknown field error")
		}
	})

	t.Run("missing database url", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, `
env: local
port: 8080
level: info
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want validation error")
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, `
port: 70000
url: postgres://localhost/toggled
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want port validation error")
		}
	})

	t.Run("invalid env", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, `
env: staging
url: postgres://localhost/toggled
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want env validation error")
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		t.Parallel()

		path := writeConfig(t, `
url: postgres://localhost/toggled
level: verbose
`)
		_, err := Load(path)
		if err == nil {
			t.Fatal("Load() error = nil, want log level error")
		}
	})
}

// TestLoadDatabaseURLOverride is kept separate and sequential: t.Setenv
// cannot be used in parallel tests.
func TestLoadDatabaseURLOverride(t *testing.T) {
	path := writeConfig(t, `
env: local
port: 9090
url: postgres://toggled:toggled@localhost:5432/toggled?sslmode=disable
level: info
`)
	override := "postgres://override:override@db.internal:5432/toggled"
	t.Setenv(DatabaseURLEnv, override)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.URL != override {
		t.Errorf("URL = %q, want %q", cfg.URL, override)
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
