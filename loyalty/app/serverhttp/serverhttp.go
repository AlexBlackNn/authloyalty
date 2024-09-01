package serverhttp

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/services/loyaltyservice"

	"github.com/AlexBlackNn/authloyalty/loyalty/cmd/router"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/handlershttp/http/v1"
)

// App service consists all entities needed to work.
type App struct {
	Cfg            *config.Config
	Log            *slog.Logger
	Srv            *http.Server
	loyaltyService *loyaltyservice.Loyalty
	HandlersV1     v1.LoyaltyHandlers
	HealthChecker  v1.HealthHandlers
}

// New creates App collecting handlers and server
func New(
	cfg *config.Config,
	log *slog.Logger,
	loyaltyService *loyaltyservice.Loyalty,
) (*App, error) {

	projectHandlersV1 := v1.New(log, cfg, loyaltyService)
	healthHandlersV1 := v1.NewHealth(log, loyaltyService)
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
