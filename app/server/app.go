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
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	patroni "github.com/AlexBlackNn/authloyalty/pkg/storage/patroni"
	redis "github.com/AlexBlackNn/authloyalty/pkg/storage/redis-sentinel"
	"google.golang.org/protobuf/proto"
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

type Sender interface {
	Send(msg proto.Message, topic string) error
}

// App service consists all entities needed to work.
type App struct {
	Cfg           *config.Config
	Log           *slog.Logger
	Srv           *http.Server
	UserStorage   UserStorage
	TokenStorage  TokenStorage
	authService   *authservice.Auth
	HandlersV1    v1.AuthHandlers
	HealthChecker v1.HealthHandlers
}

// New creates App collecting service layer, config, logger and predefined storage layer.
func New() (*App, error) {
	cfg := config.New()
	log := logger.New(cfg.Env)

	storage, err := patroni.New(cfg)
	if err != nil {
		return nil, err
	}

	kafkaURL := "localhost:9094"
	schemaRegistryURL := "http://localhost:8081"

	producer, err := broker.NewProducer(kafkaURL, schemaRegistryURL)
	tokenCache := redis.New(cfg) // Use cfg from the closure
	return NewAppInitStorage(cfg, log, storage, tokenCache, producer)
}

func NewAppInitStorage(
	cfg *config.Config,
	log *slog.Logger,
	userStorage UserStorage,
	tokenStorage TokenStorage,
	broker Sender,
) (*App, error) {

	authService := authservice.New(
		cfg,
		log,
		userStorage,
		tokenStorage,
		broker,
	)

	projectHandlersV1 := v1.New(log, authService)
	healthHandlersV1 := v1.NewHealth(log, authService)
	srv := &http.Server{
		Addr: fmt.Sprintf(cfg.Address),
		Handler: router.NewChiRouter(
			cfg,
			log,
			projectHandlersV1,
			healthHandlersV1,
		),
		ReadTimeout:  time.Duration(10) * time.Second,
		WriteTimeout: time.Duration(10) * time.Second,
		IdleTimeout:  time.Duration(10) * time.Second,
	}

	return &App{
		Cfg:           cfg,
		Log:           log,
		Srv:           srv,
		UserStorage:   userStorage,
		TokenStorage:  tokenStorage,
		authService:   authService,
		HandlersV1:    projectHandlersV1,
		HealthChecker: healthHandlersV1,
	}, nil
}