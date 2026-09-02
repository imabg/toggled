package main

import (
	"fmt"
	"os"

	"github.com/imabg/toggled/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load(config.DefaultPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "toggled: %v\n", err)
		os.Exit(1)
	}

	if _, err := config.NewLogger(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "toggled: %v\n", err)
		os.Exit(1)
	}
	defer zap.L().Sync() //nolint:errcheck

	zap.L().Info("server starting",
		zap.String("env", string(cfg.Env)),
		zap.Int("port", cfg.HTTP.Port),
	)
}
