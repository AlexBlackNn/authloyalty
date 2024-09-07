package authservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	registrationv1 "github.com/AlexBlackNn/authloyalty/commands/proto/registration.v1/registration.v1"
	"github.com/AlexBlackNn/authloyalty/sso/internal/config"
	"github.com/AlexBlackNn/authloyalty/sso/internal/domain"
	"github.com/AlexBlackNn/authloyalty/sso/internal/dto"
	jwtlib "github.com/AlexBlackNn/authloyalty/sso/internal/lib/jwt"
	"github.com/AlexBlackNn/authloyalty/sso/internal/storage"
	"github.com/AlexBlackNn/authloyalty/sso/pkg/broker"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/metadata"
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
	) (string, error)
	GetUser(
		ctx context.Context,
		uuid string,
	) (domain.User, error)
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (domain.User, error)
	UpdateSendStatus(
		ctx context.Context,
		uuid string,
		status string,
	) error
	HealthCheck(
		ctx context.Context,
	) error
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
				err := userStorage.UpdateSendStatus(
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
			err := userStorage.UpdateSendStatus(
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

const (
	TokenRevoked = 1
)

var tracer = otel.Tracer("sso service")

// HealthCheck returns service health check.
func (a *Auth) HealthCheck(ctx context.Context) error {
	log := a.log.With(
		slog.String("info", "SERVICE LAYER: HealthCheck"),
	)
	log.Info("starts getting health check")
	defer log.Info("finish getting health check")
	return a.userStorage.HealthCheck(ctx)
}

// Login logins users.
func (a *Auth) Login(
	ctx context.Context,
	reqData *dto.Login,
) (*domain.UserWithTokens, error) {
	ctx, span := tracer.Start(ctx, "service layer: login",
		trace.WithAttributes(attribute.String("handler", "login")))
	defer span.End()

	md, _ := metadata.FromIncomingContext(ctx)
	a.log.Info("span",
		"time", md.Get("timestamp"),
		"user-id", md.Get("user-id"),
		"x-trace-id", md.Get("x-trace-id"),
	)
	ctx, usrWithTokens, err := a.generateRefreshAccessToken(ctx, reqData.Email)
	if err != nil {
		a.log.Error("Generation token failed:", "err", err.Error())
		return nil, fmt.Errorf("generation token failed: %w", err)
	}
	if err = bcrypt.CompareHashAndPassword(
		usrWithTokens.PassHash, []byte(reqData.Password),
	); err != nil {
		a.log.Warn("invalid credentials")
		return nil, fmt.Errorf("invalid credentials: %w", ErrInvalidCredentials)
	}
	return usrWithTokens, nil
}

// Refresh creates new access and refresh tokens.
func (a *Auth) Refresh(
	ctx context.Context,
	reqData *dto.Refresh,
) (*domain.UserWithTokens, error) {
	ctx, span := tracer.Start(ctx, "service layer: refresh",
		trace.WithAttributes(attribute.String("handler", "refresh")))
	defer span.End()
	md, _ := metadata.FromIncomingContext(ctx)
	a.log.Info("time: %v, userId: %v", md.Get("timestamp"), md.Get("user-id"))
	log := a.log.With(
		slog.String("info", "SERVICE LAYER: auth_service.Refresh"),
		slog.String("trace-id", "trace-id from opentelemetry"),
		slog.String("user-id", "user-id from opentelemetry extracted from jwt"),
	)
	log.Info("starting validate token")
	ctx, claims, err := a.validateToken(ctx, reqData.Token)
	if err != nil {
		log.Error("token validation failed", "err", err.Error())
		return nil, fmt.Errorf("refresh: token validation failed: %w", err)
	}
	ttl := time.Duration(claims["exp"].(float64)-float64(time.Now().Unix())) * time.Second
	if claims["token_type"].(string) == "access" {
		return nil, ErrTokenWrongType
	}
	log.Info("validate token successfully")
	ctx, usrWithTokens, err := a.generateRefreshAccessToken(ctx, claims["email"].(string))
	if err != nil {
		a.log.Error("failed to generate tokens", "err", err.Error())
		return nil, err
	}
	a.log.Info("saving refresh token to redis")
	err = a.tokenStorage.SaveToken(ctx, reqData.Token, ttl)
	if err != nil {
		a.log.Error("failed to save token", "err", err.Error())
		return nil, err
	}
	a.log.Info("token saved to redis successfully")
	return usrWithTokens, nil
}

// Register registers new users.
func (a *Auth) Register(
	ctx context.Context,
	reqData *dto.Register,
) (context.Context, *domain.UserWithTokens, error) {
	const op = "SERVICE LAYER: auth_service.RegisterNewUser"

	ctx, span := tracer.Start(ctx, "service layer: register",
		trace.WithAttributes(attribute.String("handler", "register")))
	defer span.End()

	log := a.log.With(
		slog.String("trace-id", "trace-id"),
		slog.String("user-id", "user-id"),
	)
	log.Info("registering user")
	passHash, err := bcrypt.GenerateFromPassword(
		[]byte(reqData.Password), bcrypt.DefaultCost,
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(fmt.Errorf("failed to generate password: %w", err))
		log.Error("failed to generate password hash", "err", err.Error())
		return ctx, nil, fmt.Errorf("%s: %w", op, err)
	}
	// TODO: move to dto and need to add name
	uuid, err := a.userStorage.SaveUser(ctx, reqData.Email, passHash)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(fmt.Errorf("failed to save user: %w", err))
		log.Error("failed to save user", "err", err.Error())
		return ctx, nil, fmt.Errorf("%s: %w", op, err)
	}

	span.AddEvent("user registered", trace.WithAttributes(attribute.String("user-id", uuid)))
	log.Info("user registered")
	registrationMsg := registrationv1.RegistrationMessage{
		Email:    reqData.Email,
		FullName: reqData.Name,
	}
	//TODO: registration should be got from config
	ctx, err = a.producer.Send(ctx, &registrationMsg, "registration", uuid)
	if err != nil {
		// TODO: determine the err can be faced
		// No return here with err!!!, we do continue working (so-called soft degradation)
		// even kafka does not work, server is still able to process users.
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(fmt.Errorf("sending message to broker failed %w", err))
		log.Error("sending message to broker failed", "err", err.Error())
		err = a.userStorage.UpdateSendStatus(
			ctx, uuid, "failed",
		)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(attribute.Bool("error", true))
			span.RecordError(
				fmt.Errorf(
					"failed to update message status with uuid %v: %w",
					uuid, err,
				),
			)
			log.Error(
				"failed to update message status",
				"err", err.Error(),
				"uuid", uuid,
			)
		}
	}
	span.AddEvent(
		"message to broker was sent successfully",
		trace.WithAttributes(attribute.String("user-id", uuid)),
	)
	ctx, usrWithTokens, err := a.generateRefreshAccessToken(ctx, reqData.Email)
	if err != nil {
		a.log.Error("failed to generate tokens", "err", err.Error())
		return ctx, nil, err
	}
	usrWithTokens.ID = uuid
	return ctx, usrWithTokens, nil
}

// IsAdmin checks if user is admin
func (a *Auth) IsAdmin(
	ctx context.Context,
	userID string,
) (success bool, err error) {

	const op = "SERVICE LAYER: auth_service.IsAdmin"

	log := a.log.With(
		slog.String("trace-id", "trace-id"),
		slog.String("user-id", "user-id"),
	)

	log.Info("getting user from database")
	user, err := a.userStorage.GetUser(ctx, userID)
	if err != nil {
		log.Error("failed to extract user", "err", err.Error())
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user from database extracted")
	return user.IsAdmin, nil
}

// Logout revokes tokens
func (a *Auth) Logout(
	ctx context.Context,
	reqData *dto.Logout,
) (success bool, err error) {

	log := a.log.With(
		slog.String("info", "SERVICE LAYER: auth_service.Logout"),
		slog.String("trace-id", "trace-id from opentelemetry"),
		slog.String("user-id", "user-id from opentelemetry extracted from jwt"),
	)

	log.Info("starting validate token")
	ctx, claims, err := a.validateToken(ctx, reqData.Token)
	if err != nil {
		log.Error("failed validate token: ", "err", err.Error())
		return false, err
	}
	ttl := time.Duration(claims["exp"].(float64)-float64(time.Now().Unix())) * time.Second

	log.Info("validate token successfully")
	log.Info("saving token to redis")

	err = a.tokenStorage.SaveToken(ctx, reqData.Token, ttl)
	if err != nil {
		log.Error("failed to save token", "err", err.Error())
		return false, err
	}
	log.Info("token saved to redis successfully")
	return true, nil
}

// Validate validates tokens
func (a *Auth) Validate(
	ctx context.Context,
	token string,
) (success bool, err error) {

	log := a.log.With(
		slog.String("info", "SERVICE LAYER: auth_service.Verify"),
		slog.String("trace-id", "trace-id from opentelemetry"),
		slog.String("user-id", "user-id from opentelemetry extracted from jwt"),
	)
	log.Info("starting validate token")
	ctx, _, err = a.validateToken(ctx, token)
	if err != nil {
		log.Info("failed validate token: ", err.Error())
		return false, err
	}
	log.Info("validate token successfully")
	return true, nil
}

func (a *Auth) validateToken(ctx context.Context, token string) (context.Context, jwt.MapClaims, error) {

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(a.cfg.ServiceSecret), nil
	})
	if err != nil {
		return ctx, jwt.MapClaims{}, ErrTokenParsing
	}
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return ctx, jwt.MapClaims{}, ErrTokenParsing
	}
	// check ttl
	ttl := time.Duration(claims["exp"].(float64)-float64(time.Now().Unix())) * time.Second
	if ttl < 0 {
		return ctx, jwt.MapClaims{}, ErrTokenTTLExpired
	}
	// check type of token
	if (claims["token_type"] != "refresh") && claims["token_type"] != "access" {
		return ctx, jwt.MapClaims{}, ErrTokenWrongType
	}
	// check if token exists in redis

	value, err := a.tokenStorage.CheckTokenExists(ctx, token)
	if err != nil {
		return ctx, jwt.MapClaims{}, fmt.Errorf("validateToken: %w", err)
	}
	if value == TokenRevoked {
		return ctx, jwt.MapClaims{}, ErrTokenRevoked
	}
	return ctx, claims, nil
}

func (a *Auth) generateRefreshAccessToken(
	ctx context.Context,
	email string,
) (context.Context, *domain.UserWithTokens, error) {

	user, err := a.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return ctx, nil, ErrInvalidCredentials
		}
		return ctx, nil, err
	}

	accessToken, err := jwtlib.NewToken(user, a.cfg, "access")
	if err != nil {
		return ctx, nil, fmt.Errorf("accessToken generation failed: %w", err)
	}
	refreshToken, err := jwtlib.NewToken(user, a.cfg, "refresh")
	if err != nil {
		return ctx, nil, fmt.Errorf("refreshToken generation failed: %w", err)
	}
	return ctx, &domain.UserWithTokens{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
