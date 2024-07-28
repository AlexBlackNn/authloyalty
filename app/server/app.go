package server

import (
	"context"
	"fmt"
	"github.com/AlexBlackNn/metrics/cmd/server/router"
	"github.com/AlexBlackNn/metrics/internal/config/configserver"
	"github.com/AlexBlackNn/metrics/internal/domain/models"
	"github.com/AlexBlackNn/metrics/internal/handlers/v1"
	v2 "github.com/AlexBlackNn/metrics/internal/handlers/v2"
	v3 "github.com/AlexBlackNn/metrics/internal/handlers/v3"
	"github.com/AlexBlackNn/metrics/internal/logger"
	"github.com/AlexBlackNn/metrics/internal/migrator"
	"github.com/AlexBlackNn/metrics/internal/services/metricsservice"
	"github.com/AlexBlackNn/metrics/pkg/storage/memstorage"
	"github.com/AlexBlackNn/metrics/pkg/storage/postgres"
	"log/slog"
	"net/http"
	"time"
)

type MetricsStorage interface {
	UpdateMetric(
		ctx context.Context,
		metric models.MetricGetter,
	) error
	UpdateSeveralMetrics(
		ctx context.Context,
		metrics map[string]models.MetricGetter,
	) error
	GetMetric(
		ctx context.Context,
		metric models.MetricGetter,
	) (models.MetricGetter, error)
	GetAllMetrics(
		ctx context.Context,
	) ([]models.MetricGetter, error)
}

type HealthChecker interface {
	HealthCheck(
		ctx context.Context,
	) error
}

// App service consists all entities needed to work.
type App struct {
	MetricsService   *metricsservice.MetricService
	HandlersV1       v1.MetricHandlers
	HandlersV2       v2.MetricHandlers
	HealthHandlersV2 v2.HealthHandlers
	HandlersV3       v3.MetricHandlers
	Cfg              *configserver.Config
	Log              *slog.Logger
	Srv              *http.Server
	DataBase         MetricsStorage
	HealthChecker    HealthChecker
}

// New creates App collecting service layer, config, logger and predefined storage layer.
func New() (*App, error) {

	cfg, err := configserver.New()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.Env)

	// Err is now skipped, but when migratings to postgres/sqlite/etc... err will be checked.
	if cfg.ServerDataBaseDSN != "" {
		postgresStorage, err := postgres.New(cfg, log)
		if err != nil {
			return nil, err
		}
		log.Info("Starts to apply migrations")
		err = migrator.ApplyMigration(cfg)
		if err != nil {
			log.Error("Failed to apply migration", "err", err.Error())
		}
		log.Info("Finish to apply migrations")
		return NewAppInitStorage(postgresStorage, postgresStorage, cfg, log)
	}
	memStorage, err := memstorage.New(cfg, log)
	if err != nil {
		return nil, err
	}
	return NewAppInitStorage(memStorage, memStorage, cfg, log)
}

func NewAppInitStorage(ms MetricsStorage, hc HealthChecker, cfg *configserver.Config, log *slog.Logger) (*App, error) {

	metricsService := metricsservice.New(
		log,
		cfg,
		ms,
		hc,
	)

	projectHandlersV1 := v1.New(log, metricsService)
	projectHandlersV2 := v2.New(log, metricsService)
	healthHandlersV2 := v2.NewHealth(log, metricsService)
	projectHandlersV3 := v3.New(log, metricsService)

	srv := &http.Server{
		Addr: fmt.Sprintf(cfg.ServerAddr),
		Handler: router.NewChiRouter(
			cfg,
			log,
			projectHandlersV1,
			projectHandlersV2,
			healthHandlersV2,
			projectHandlersV3,
		),
		ReadTimeout:  time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.ServerIdleTimeout) * time.Second,
	}

	return &App{
		MetricsService:   metricsService,
		HandlersV1:       projectHandlersV1,
		HandlersV2:       projectHandlersV2,
		HealthHandlersV2: healthHandlersV2,
		HandlersV3:       projectHandlersV3,
		Srv:              srv,
		Cfg:              cfg,
		Log:              log,
		DataBase:         ms,
		HealthChecker:    hc,
	}, nil
}
