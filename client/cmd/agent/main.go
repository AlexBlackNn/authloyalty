package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/AlexBlackNn/authloyalty/client/app/agent"
	"github.com/AlexBlackNn/authloyalty/client/internal/config/configagent"
	"github.com/AlexBlackNn/authloyalty/client/internal/logger"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	var wg sync.WaitGroup

	cfg, err := configagent.New()
	if err != nil {
		log.Fatalln(err)
	}

	log_ := logger.New(cfg.Env)
	showProjectInfo(log_)
	log_.Info("starting application", slog.String("env", cfg.Env))

	appMonitor, err := agent.NewAppMonitor(log_, cfg)
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	wg.Add(4)
	go func() {
		defer wg.Done()
		<-stop
		cancel()
	}()

	go func() {
		defer wg.Done()
		appMonitor.MetricsService.Collect(ctx)
	}()

	go func() {
		defer wg.Done()
		appMonitor.MetricsService.CollectAddition(ctx)
	}()

	go func() {
		defer wg.Done()
		appMonitor.MetricsService.Send(ctx)
	}()

	wg.Wait()
}

func showProjectInfo(log *slog.Logger) {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	projInfo := fmt.Sprintf(
		"Build version: %s, Build date: %s, Build commit: %s",
		buildVersion, buildDate, buildCommit,
	)
	log.Info(projInfo)
}
