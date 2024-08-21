package serverhttp

import (
	"fmt"
	"github.com/AlexBlackNn/authloyalty/cmd/sso/router"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/handlershttp/http_v1"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	"log/slog"
	"net/http"
	"time"
)

// App service consists all entities needed to work.
type App struct {
	Cfg           *config.Config
	Log           *slog.Logger
	Srv           *http.Server
	authService   *authservice.Auth
	HandlersV1    http_v1.AuthHandlers
	HealthChecker http_v1.HealthHandlers
}

// New creates App collecting handlers and server
func New(
	cfg *config.Config,
	log *slog.Logger,
	authService *authservice.Auth,
) (*App, error) {

	projectHandlersV1 := http_v1.New(log, cfg, authService)
	healthHandlersV1 := http_v1.NewHealth(log, authService)
	srv := &http.Server{
		Addr: fmt.Sprintf(cfg.Address),
		Handler: router.NewChiRouter(
			cfg,
			log,
			projectHandlersV1,
			healthHandlersV1,
		),
		ReadTimeout:  time.Duration(cfg.ServerTimeout.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerTimeout.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.ServerTimeout.IdleTimeout) * time.Second,
	}
	return &App{
		Cfg:           cfg,
		Log:           log,
		Srv:           srv,
		HandlersV1:    projectHandlersV1,
		HealthChecker: healthHandlersV1,
	}, nil
}
