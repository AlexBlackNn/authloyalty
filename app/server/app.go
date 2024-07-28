package server

import (
	"context"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/cmd/router"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	v1 "github.com/AlexBlackNn/authloyalty/internal/handlers/v1"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	authservice "github.com/AlexBlackNn/authloyalty/internal/services/auth_service"
	"log/slog"
	"net/http"
	"time"
)

type UserStorage interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (context.Context, int64, error)
	GetUser(
		ctx context.Context,
		value any,
	) (context.Context, models.User, error)
}

type TokenStorage interface {
	SaveToken(ctx context.Context, token string, ttl time.Duration) (context.Context, error)
	GetToken(ctx context.Context, token string) (context.Context, string, error)
	CheckTokenExists(ctx context.Context, token string) (context.Context, int64, error)
}

type HealthChecker interface {
	HealthCheck(
		ctx context.Context,
	) error
}

// App service consists all entities needed to work.
type App struct {
	MetricsService *authservice.Auth
	HandlersV1     v1.AuthHandlers
	Cfg            *config.Config
	Log            *slog.Logger
	Srv            *http.Server
	DataBase       MetricsStorage
	HealthChecker  v1.HealthHandlers
}

// New creates App collecting service layer, config, logger and predefined storage layer.
func New() (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}
	log := logger.New(cfg.Env)
	return NewAppInitStorage(postgresStorage, postgresStorage, cfg, log)

}

func NewAppInitStorage(ms MetricsStorage, hc HealthChecker, cfg *configserver.Config, log *slog.Logger) (*App, error) {

	metricsService := metricsservice.New(
		log,
		cfg,
		ms,
		hc,
	)

	projectHandlersV1 := v1.New(log, metricsService)

	srv := &http.Server{
		Addr: fmt.Sprintf(cfg.ServerAddr),
		Handler: router.NewChiRouter(
			cfg,
			log,
			projectHandlersV1,
		),
		ReadTimeout:  time.Duration(10) * time.Second,
		WriteTimeout: time.Duration(10) * time.Second,
		IdleTimeout:  time.Duration(10) * time.Second,
	}

	return &App{
		MetricsService: metricsService,
		HandlersV1:     projectHandlersV1,
		Srv:            srv,
		Cfg:            cfg,
		Log:            log,
		DataBase:       ms,
		HealthChecker:  hc,
	}, nil
}
