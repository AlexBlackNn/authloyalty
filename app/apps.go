package app

import (
	"context"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/app/servergrpc"
	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	authservice "github.com/AlexBlackNn/authloyalty/internal/services/auth_service"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	patroni "github.com/AlexBlackNn/authloyalty/pkg/storage/patroni"
	redis "github.com/AlexBlackNn/authloyalty/pkg/storage/redissentinel"
	"github.com/prometheus/common/log"
	"google.golang.org/protobuf/proto"
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
	Stop() error
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

type Sender interface {
	Send(msg proto.Message, topic string, key string) error
}

type App struct {
	ServerHttp         *serverhttp.App
	ServerGrpc         *servergrpc.App
	ServerUserStorage  UserStorage
	ServerTokenStorage TokenStorage
	ServerProducer     Sender
}

func (a *App) MustStart() {
	go func() {
		log.Info("http server starting")
		if err := a.ServerHttp.Srv.ListenAndServe(); err != nil {
			panic(err)
		}
		log.Info("http server started successfully")
	}()

	log.Info("grpc server starting")
	a.ServerGrpc.Srv.Bootstrap(context.Background())
	log.Info("grpc server started successfully")
}

func (a *App) Stop() error {
	err := a.ServerUserStorage.Stop()
	if err != nil {
		return err
	}
	//TODO: add other entities closure
	return nil
}

func New() (*App, error) {

	cfg := config.New()
	log := logger.New(cfg.Env)

	userStorage, err := patroni.New(cfg)
	if err != nil {
		return nil, err
	}

	tokenStorage := redis.New(cfg)

	producer, err := broker.NewProducer(cfg.Kafka.KafkaURL, cfg.Kafka.SchemaRegistryURL)

	go func() {
		for kafkaResponse := range producer.ResponseChan {
			fmt.Println("http", kafkaResponse)
		}
	}()

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
	serverGrpc, err := servergrpc.New(cfg, log, authService)
	if err != nil {
		return nil, err
	}
	return &App{
		ServerHttp: serverHttp,
		ServerGrpc: serverGrpc,
	}, nil
}
