package config

import (
	"slices"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestBuildZapConfig(t *testing.T) {
	t.Parallel()

	t.Run("production info uses json encoding on stdout", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Env:   EnvProduction,
			Level: "info",
		}
		zapCfg, err := buildZapConfig(cfg)
		if err != nil {
			t.Fatalf("buildZapConfig() unexpected error: %v", err)
		}
		if zapCfg.Encoding != "json" {
			t.Errorf("Encoding = %q, want json", zapCfg.Encoding)
		}
		if !slices.Contains(zapCfg.OutputPaths, "stdout") {
			t.Errorf("OutputPaths = %v, want stdout", zapCfg.OutputPaths)
		}
		if !zapCfg.Level.Enabled(zapcore.InfoLevel) {
			t.Error("info level should be enabled")
		}
		if zapCfg.Level.Enabled(zapcore.DebugLevel) {
			t.Error("debug level should be disabled at info")
		}
	})

	t.Run("local env uses console encoding", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Env:   EnvLocal,
			Level: "info",
		}
		zapCfg, err := buildZapConfig(cfg)
		if err != nil {
			t.Fatalf("buildZapConfig() unexpected error: %v", err)
		}
		if zapCfg.Encoding != "console" {
			t.Errorf("Encoding = %q, want console", zapCfg.Encoding)
		}
	})

	t.Run("debug level uses console encoding even in production", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Env:   EnvProduction,
			Level: "debug",
		}
		zapCfg, err := buildZapConfig(cfg)
		if err != nil {
			t.Fatalf("buildZapConfig() unexpected error: %v", err)
		}
		if zapCfg.Encoding != "console" {
			t.Errorf("Encoding = %q, want console", zapCfg.Encoding)
		}
		if !zapCfg.Level.Enabled(zapcore.DebugLevel) {
			t.Error("debug level should be enabled")
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Env:   EnvProduction,
			Level: "verbose",
		}
		_, err := buildZapConfig(cfg)
		if err == nil {
			t.Fatal("buildZapConfig() error = nil, want error")
		}
	})
}

// TestNewLogger is kept sequential: NewLogger installs the logger as the
// process-wide zap global, which is unsafe to run in parallel.
func TestNewLogger(t *testing.T) {
	t.Cleanup(func() {
		zap.ReplaceGlobals(zap.NewNop())
	})

	before := zap.L()
	cfg := &Config{
		Env:   EnvProduction,
		Level: "info",
	}
	if err := NewLogger(cfg); err != nil {
		t.Fatalf("NewLogger() unexpected error: %v", err)
	}
	if zap.L() == before {
		t.Error("NewLogger() did not install the logger as zap.L()")
	}
	_ = zap.L().Sync()
}
