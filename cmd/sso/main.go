package main

import (
	"github.com/AlexBlackNn/authloyalty/app"
	"github.com/AlexBlackNn/authloyalty/app/server"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// http
	application, err := server.New()
	if err != nil {
		panic(err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	//application.Log.Info("starting application", slog.String("cfg", application.Cfg.String()))
	go func() {
		if err = application.Srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	application.Log.Info("http server started")

	// GRPC
	// init logger
	log := logger.New(application.Cfg.Env)
	log.Info("starting application", slog.String("env", application.Cfg.Env))
	// init app
	app.New(log, application.Cfg)

	signalType := <-stop
	application.Log.Info(
		"application stopped",
		slog.String("signalType",
			signalType.String()),
	)

}
