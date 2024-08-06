package main

import (
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
	application.MustStart()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	signalType := <-stop

	err = application.Stop()
	if err != nil {
		log.Error(
			"server failed to stop",
			"err", err.Error(),
			slog.String("signalType", signalType.String()),
		)
		return
	}
	log.Info(
		"application stopped",
		slog.String("signalType", signalType.String()),
	)
}
