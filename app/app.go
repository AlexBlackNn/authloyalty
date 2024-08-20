package app

import (
	"context"
	"errors"
	"github.com/AlexBlackNn/authloyalty/app/servergrpc"
	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	"github.com/AlexBlackNn/authloyalty/pkg/storage"
	"github.com/AlexBlackNn/authloyalty/pkg/storage/patroni"
	"github.com/AlexBlackNn/authloyalty/pkg/storage/redissentinel"
	"github.com/AlexBlackNn/authloyalty/pkg/tracing"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"
	log "log/slog"
	"time"
)

type UserStorage interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (context.Context, string, error)
	GetUser(
		ctx context.Context,
		email string,
	) (context.Context, models.User, error)
	Stop() error
}

type TokenStorage interface {
	SaveToken(
		ctx context.Context,
		token string,
		ttl time.Duration,
	) (context.Context, error)
	GetToken(
		ctx context.Context,
		token string,
	) (context.Context, string, error)
	CheckTokenExists(
		ctx context.Context,
		token string,
	) (context.Context, int64, error)
}

type HealthChecker interface {
	HealthCheck(
		ctx context.Context,
	) (context.Context, error)
}

type SendCloser interface {
	Send(
		ctx context.Context,
		msg proto.Message,
		topic string,
		key string,
	) (context.Context, error)
	Close()
}

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

type App struct {
	ServerHttp          *serverhttp.App
	ServerGrpc          *servergrpc.App
	ServerUserStorage   UserStorage
	ServerTokenStorage  TokenStorage
	ServerProducer      SendCloser
	ServerOpenTelemetry *trace.TracerProvider
}

func (a *App) startHttpServer() chan error {
	errChan := make(chan error)
	go func() {
		if err := a.ServerHttp.Srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

func (a *App) Start(ctx context.Context) error {
	log.Info("grpc server starting")
	a.ServerGrpc.Srv.Bootstrap(context.Background())
	log.Info("http server starting")
	errChan := a.startHttpServer()
	select {
	case <-ctx.Done():
		return a.Stop()
	case err := <-errChan:
		return err
	}
}

// TODO: close all unexported methods

func (a *App) Stop() error {
	log.Info("close user storage client")
	err := a.ServerUserStorage.Stop()
	if err != nil {
		return err
	}
	log.Info("close information bus client")
	a.ServerProducer.Close()

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

func New() (*App, error) {

	cfg := config.New()
	log := logger.New(cfg.Env)

	userStorage, err := patroni.New(cfg)
	if err != nil {
		if !errors.Is(err, storage.ErrConnection) {
			return nil, err
		}
		log.Warn(err.Error())
	}

	tokenStorage, err := redissentinel.New(cfg)
	if err != nil {
		return nil, err
	}

	producer, err := broker.New(cfg)
	if err != nil {
		return nil, err
	}

	authService := authservice.New(
		cfg,
		log,
		userStorage,
		tokenStorage,
		producer,
	)

	// http server
	serverHttp, err := serverhttp.New(cfg, log, authService)
	if err != nil {
		return nil, err
	}

	// grpc server
	serverGrpc, err := servergrpc.New(authService)
	if err != nil {
		return nil, err
	}

	tp, err := tracing.Init("sso service", cfg)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &App{
		ServerHttp:          serverHttp,
		ServerGrpc:          serverGrpc,
		ServerUserStorage:   userStorage,
		ServerTokenStorage:  tokenStorage,
		ServerProducer:      producer,
		ServerOpenTelemetry: tp,
	}, nil
}
