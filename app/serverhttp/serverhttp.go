package serverhttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/cmd/router"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	handlers "github.com/AlexBlackNn/authloyalty/internal/handlers/v1"
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

// App service consists all entities needed to work.
type App struct {
	Cfg           *config.Config
	Log           *slog.Logger
	Srv           *http.Server
	UserStorage   UserStorage
	TokenStorage  TokenStorage
	authService   *authservice.Auth
	HandlersV1    handlers.AuthHandlers
	HealthChecker handlers.HealthHandlers
}

// New creates App collecting service layer, config, logger and predefined storage layer.
func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	// TODO: seems to need factory here
	//storage, err := postgres.New(cfg.StoragePath) //uncomment for postgres
	userStorage, err := patroni.New(cfg)
	if err != nil {
		return nil, err
	}

	tokenStorage := redis.New(cfg)

	producer, kafkaResponseChan, err := broker.NewProducer(cfg.Kafka.KafkaURL, cfg.Kafka.SchemaRegistryURL)

	go func() {
		for kafkaResponse := range kafkaResponseChan {
			fmt.Println(kafkaResponse)
		}
	}()

	if err != nil {
		return nil, err
	}
	return NewAppInitStorage(cfg, log, userStorage, tokenStorage, producer)
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

	projectHandlersV1 := handlers.New(log, authService)
	healthHandlersV1 := handlers.NewHealth(log, authService)
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

// TODO: panic happens panic: http: Server closed if ctrl +C think why???

func (a *App) Stop() error {
	var errs error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Shutdown the server and handle the timeout error immediately
	if err := a.Srv.Shutdown(ctx); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			errs = errors.Join(errs, err)
		}
	}
	// Stop the user storage and join any errors to the `errs` variable
	if err := a.UserStorage.Stop(); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}
