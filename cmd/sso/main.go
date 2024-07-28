package main

import (
	"github.com/AlexBlackNn/authloyalty/app"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"log/slog"
)

func main() {
	// init config
	cfg := config.MustLoad()
	// init logger
	log := logger.New(cfg.Env)
	log.Info("starting application", slog.String("env", cfg.Env))
	// init app
	app.New(log, cfg)
}
