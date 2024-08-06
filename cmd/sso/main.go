package main

import (
	"github.com/AlexBlackNn/authloyalty/app"
	"github.com/prometheus/common/log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// @title           Swagger API
// @version         1.0
// @description     sso service.
// @contact.name   API Support
// @license.name  Apache 2.0
// @license.calculation   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8000
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
//
//go:generate go run github.com/swaggo/swag/cmd/swag init
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
