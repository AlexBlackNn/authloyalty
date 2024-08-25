package loyaltyservice

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/sso/pkg/broker"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

type getResponseChanSender interface {
	Send(
		ctx context.Context,
		msg proto.Message,
		topic string,
		key string,
	) (context.Context, error)
	GetResponseChan() chan *broker.Response
}

type userStorage interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (context.Context, string, error)
	GetUser(
		ctx context.Context,
		uuid string,
	) (context.Context, domain.User, error)
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (context.Context, domain.User, error)
	UpdateSendStatus(
		ctx context.Context,
		uuid string,
		status string,
	) (context.Context, error)
	HealthCheck(
		ctx context.Context,
	) (context.Context, error)
}

type tokenStorage interface {
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

type Auth struct {
	log          *slog.Logger
	userStorage  userStorage
	tokenStorage tokenStorage
	producer     getResponseChanSender
	cfg          *config.Config
}

// New returns a new instance of Auth service
func New(
	cfg *config.Config,
	log *slog.Logger,
	userStorage userStorage,
	tokenStorage tokenStorage,
	producer getResponseChanSender,
) *Auth {
	// Channel that is used by kafka to return sent message status.
	brokerRespChan := producer.GetResponseChan()
	// Getting status (async) from channel to determine if a message was sent successfully.
	go func() {
		for brokerResponse := range brokerRespChan {
			if brokerResponse.Err != nil {
				if errors.Is(brokerResponse.Err, broker.KafkaError) {
					log.Error("broker error", "err", brokerResponse.Err)
					continue
				}
				log.Error(
					"broker response error on message",
					"err", brokerResponse.Err,
					"uuid", brokerResponse.UserUUID,
				)
				_, err := userStorage.UpdateSendStatus(
					context.Background(), brokerResponse.UserUUID, "failed",
				)
				if err != nil {
					log.Error(
						"failed to update message status",
						"err", err.Error(),
						"uuid", brokerResponse.UserUUID,
					)
				}
				continue
			}
			_, err := userStorage.UpdateSendStatus(
				context.Background(),
				brokerResponse.UserUUID,
				"successful",
			)
			if err != nil {
				log.Error("failed to update message status", "err", err.Error())
			}
		}
	}()

	return &Auth{
		log:          log,
		userStorage:  userStorage,
		tokenStorage: tokenStorage,
		producer:     producer,
		cfg:          cfg,
	}
}

var tracer = otel.Tracer("sso service")

// HealthCheck returns service health check.
func (a *Auth) HealthCheck(ctx context.Context) (context.Context, error) {
	log := a.log.With(
		slog.String("info", "SERVICE LAYER: metrics_service.HealthCheck"),
	)
	log.Info("starts getting health check")
	defer log.Info("finish getting health check")
	return a.userStorage.HealthCheck(ctx)
}

// Logout revokes tokens
func (a *Auth) Logout(
	ctx context.Context,
	,
) (success bool, err error) {

}

// Validate validates tokens
func (a *Auth) Validate(
	ctx context.Context,
	token string,
) (success bool, err error) {

}
