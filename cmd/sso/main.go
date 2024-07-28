package main

import (
	"github.com/AlexBlackNn/authloyalty/app/server"
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
	application.Log.Info("server started")

	//// GRPC
	//// init config
	//cfg := config.New()
	//// init logger
	//log := logger.New(cfg.Env)
	//log.Info("starting application", slog.String("env", cfg.Env))
	//// init app
	//app.New(log, cfg)

	signalType := <-stop
	application.Log.Info(
		"application stopped",
		slog.String("signalType",
			signalType.String()),
	)

}
