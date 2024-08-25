package app

import (
	"context"
	"errors"
	"github.com/AlexBlackNn/authloyalty/loyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/services/loyaltyservice"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/storage"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/storage/patroni"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/tracing"
	"github.com/AlexBlackNn/authloyalty/sso/app/servergrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	log "log/slog"
)

type loyaltyStorage interface {
	AddLoyalty(
		ctx context.Context,
		userLoyalty domain.UserLoyalty,
	) (context.Context, domain.UserLoyalty, error)
	SubLoyalty(
		ctx context.Context,
		userLoyalty domain.UserLoyalty,
	) (context.Context, domain.UserLoyalty, error)
	GetLoyalty(
		ctx context.Context,
		userLoyalty domain.UserLoyalty,
	) (context.Context, domain.UserLoyalty, error)
	Stop() error
}

type App struct {
	ServerHttp           *serverhttp.App
	ServerGrpc           *servergrpc.App
	ServerLoyaltyStorage loyaltyStorage
	ServerOpenTelemetry  *trace.TracerProvider
}

func New() (*App, error) {
	cfg := config.New()
	log := logger.New(cfg.Env)

	loyalStorage, err := patroni.New(cfg)
	if err != nil {
		if !errors.Is(err, storage.ErrConnection) {
			return nil, err
		}
		log.Warn(err.Error())
	}

	authService := loyaltyservice.New(
		cfg,
		log,
		loyalStorage,
	)

	// http server
	serverHttp, err := serverhttp.New(cfg, log, authService)
	if err != nil {
		return nil, err
	}

	tp, err := tracing.Init("sso service", cfg)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &App{
		ServerHttp:           serverHttp,
		ServerLoyaltyStorage: loyalStorage,
		ServerOpenTelemetry:  tp,
	}, nil
}

func (a *App) startHTTPServer() chan error {
	errChan := make(chan error)
	go func() {
		if err := a.ServerHttp.Srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

func (a *App) Start(ctx context.Context) error {
	log.Info("http server starting")
	errHTTPChan := a.startHTTPServer()
	select {
	case <-ctx.Done():
		return a.Stop()
	case httpErr := <-errHTTPChan:
		return httpErr
	}
}

// TODO: close all unexported methods
func (a *App) Stop() error {
	log.Info("close user storage client")
	err := a.ServerLoyaltyStorage.Stop()
	if err != nil {
		return err
	}

	log.Info("close http server")
	err = a.ServerHttp.Srv.Close()
	if err != nil {
		return err
	}

	log.Info("close open telemetry client")
	err = a.ServerOpenTelemetry.Shutdown(context.Background())
	if err != nil {
		return err
	}
	return nil
}
