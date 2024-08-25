package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/AlexBlackNn/authloyalty/sso/app"
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
		log.Fatal(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	err = application.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
