package main

import (
	"context"
	"github.com/AlexBlackNn/authloyalty/app/servergrpc"
	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg.Env)

	// http
	serverHttp, err := serverhttp.New(cfg, log)
	if err != nil {
		panic(err)
	}

	go func() {
		if err = serverHttp.Srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	log.Info("http server started")

	// grpc
	serverGrpc, err := servergrpc.New(cfg, log)
	if err != nil {
		panic(err)
	}
	serverGrpc.Srv.Bootstrap(context.Background())
	log.Info("grpc  server started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	signalType := <-stop
	log.Info(
		"application stopped",
		slog.String("signalType",
			signalType.String()),
	)

}
