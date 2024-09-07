package app

import (
	"context"
	"errors"
	log "log/slog"
	"time"

	"github.com/AlexBlackNn/authloyalty/sso/app/servergrpc"
	"github.com/AlexBlackNn/authloyalty/sso/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/sso/internal/config"
	"github.com/AlexBlackNn/authloyalty/sso/internal/domain"
	"github.com/AlexBlackNn/authloyalty/sso/internal/logger"
	"github.com/AlexBlackNn/authloyalty/sso/internal/services/authservice"
	"github.com/AlexBlackNn/authloyalty/sso/internal/storage"
	"github.com/AlexBlackNn/authloyalty/sso/internal/storage/patroni"
	"github.com/AlexBlackNn/authloyalty/sso/internal/storage/redissentinel"
	"github.com/AlexBlackNn/authloyalty/sso/pkg/broker"
	"github.com/AlexBlackNn/authloyalty/sso/pkg/tracing"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"
)

type userStorage interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (string, error)
	GetUser(
		ctx context.Context,
		email string,
	) (domain.User, error)
	Stop() error
}

type tokenStorage interface {
	SaveToken(
		ctx context.Context,
		token string,
		ttl time.Duration,
	) error
	GetToken(
		ctx context.Context,
		token string,
	) (string, error)
	CheckTokenExists(
		ctx context.Context,
		token string,
	) (int64, error)
}

type sendCloser interface {
	Send(
		ctx context.Context,
		msg proto.Message,
		topic string,
		key string,
	) error
	Close()
}

type App struct {
	ServerHttp          *serverhttp.App
	ServerGrpc          *servergrpc.App
	ServerUserStorage   userStorage
	ServerTokenStorage  tokenStorage
	ServerProducer      sendCloser
	ServerOpenTelemetry *trace.TracerProvider
}

func New() (*App, error) {

	cfg := config.New()
	log := logger.New(cfg.Env)

	usrStorage, err := patroni.New(cfg)
	if err != nil {
		if !errors.Is(err, storage.ErrConnection) {
			return nil, err
		}
		log.Warn(err.Error())
	}

	tknStorage, err := redissentinel.New(cfg)
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
		usrStorage,
		tknStorage,
		producer,
	)

	// http server
	serverHttp, err := serverhttp.New(cfg, log, authService)
	if err != nil {
		return nil, err
	}

	// grpc server
	serverGrpc, err := servergrpc.New(cfg, log, authService)
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
		ServerUserStorage:   usrStorage,
		ServerTokenStorage:  tknStorage,
		ServerProducer:      producer,
		ServerOpenTelemetry: tp,
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

func (a *App) startGRPCServer() chan error {
	errChan := make(chan error)
	go func() {
		if err := a.ServerGrpc.Start(); err != nil {
			errChan <- err
		}
	}()
	return errChan
}

func (a *App) Start(ctx context.Context) error {
	log.Info("grpc server starting")
	errGRPCChan := a.startGRPCServer()
	log.Info("http server starting")
	errHTTPChan := a.startHTTPServer()
	select {
	case <-ctx.Done():
		return a.Stop()
	case httpErr := <-errHTTPChan:
		return httpErr
	case grpcErr := <-errGRPCChan:
		return grpcErr
	}
}

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

	log.Info("close grpc server")
	a.ServerGrpc.Server.Stop()

	log.Info("close open telemetry client")
	err = a.ServerOpenTelemetry.Shutdown(context.Background())
	if err != nil {
		return err
	}
	return nil
}
