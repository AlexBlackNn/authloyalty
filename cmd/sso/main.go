package main

import (
	"context"
	"github.com/AlexBlackNn/authloyalty/app"
	"github.com/prometheus/common/log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	application, err := app.New()
	if err != nil {
		panic(err)
	}
	log.Info("http server starting")
	go func() {
		if err = application.ServerHttp.Srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	log.Info("http server started successfully")

	log.Info("grpc server starting")
	application.ServerGrpc.Srv.Bootstrap(context.Background())
	log.Info("grpc server started successfully")

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	signalType := <-stop

	err = application.ServerHttp.Stop()
	if err != nil {
		log.Error("http server failed to stop", "err", err.Error(), "signal", signalType)
	}
	err = application.ServerGrpc.Stop()
	if err != nil {
		log.Error("grpc server failed to stop", "err", err.Error(), "signal", signalType)
	}
	log.Info(
		"application stopped",
		slog.String("signalType",
			signalType.String()),
	)
}
