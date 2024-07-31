package main

import (
	"fmt"
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
	serverhttp, err := serverhttp.New(cfg, log)
	if err != nil {
		panic(err)
	}

	go func() {
		if err = serverhttp.Srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	serverhttp.Log.Info("http server started")

	// grpc
	_, err = servergrpc.New(cfg, log)
	if err != nil {
		panic(err)
	}
	fmt.Println("11111111111111111111111111111111")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	serverhttp.Log.Info("grpc server started")

	signalType := <-stop
	log.Info(
		"application stopped",
		slog.String("signalType",
			signalType.String()),
	)

}
