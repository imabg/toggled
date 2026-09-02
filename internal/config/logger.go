package config

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger builds a zap logger from cfg, writes to stdout, and installs
// it as the process-wide default (zap.L / zap.S).
// Production uses JSON encoding. Local env or debug level uses a
// human-readable console encoder.
func NewLogger(cfg *Config) (*zap.Logger, error) {
	zapCfg, err := buildZapConfig(cfg)
	if err != nil {
		return nil, err
	}
	log, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("config: build logger: %w", err)
	}
	zap.ReplaceGlobals(log)
	return log, nil
}

func buildZapConfig(cfg *Config) (zap.Config, error) {
	level, err := parseLevel(cfg.Log.Level)
	if err != nil {
		return zap.Config{}, err
	}

	useConsole := cfg.Env == EnvLocal || level == zapcore.DebugLevel
	var zapCfg zap.Config
	if useConsole {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}
	return zapCfg, nil
}

func parseLevel(level string) (zapcore.Level, error) {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(strings.ToLower(level))); err != nil {
		return 0, fmt.Errorf("config: log.level must be debug, info, warn, or error")
	}
	return l, nil
}
