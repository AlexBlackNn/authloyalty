package auth_service

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	jwtlib "github.com/AlexBlackNn/authloyalty/internal/lib/jwt"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	storage2 "github.com/AlexBlackNn/authloyalty/pkg/storage"
	"github.com/AlexBlackNn/authloyalty/protos/proto/registration/registration.v1"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"time"
)

type GetResponseChanSender interface {
	Send(msg proto.Message, topic string, key string) error
	GetResponseChan() chan *broker.Response
}

type UserStorage interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (context.Context, string, error)
	GetUser(
		ctx context.Context,
		uuid string,
	) (context.Context, models.User, error)
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (context.Context, models.User, error)
	UpdateSendStatus(ctx context.Context, uuid string, status string) (context.Context, error)
}

type TokenStorage interface {
	SaveToken(ctx context.Context, token string, ttl time.Duration) (context.Context, error)
	GetToken(ctx context.Context, token string) (context.Context, string, error)
	CheckTokenExists(ctx context.Context, token string) (context.Context, int64, error)
}

type Auth struct {
	log          *slog.Logger
	userStorage  UserStorage
	tokenStorage TokenStorage
	producer     GetResponseChanSender
	cfg          *config.Config
}

// New returns a new instance of Auth service
func New(
	cfg *config.Config,
	log *slog.Logger,
	userStorage UserStorage,
	tokenStorage TokenStorage,
	producer GetResponseChanSender,
) *Auth {
	brokerRespChan := producer.GetResponseChan()

	go func() {
		for brokerResponse := range brokerRespChan {
			log.Debug("broker response", "resp", brokerResponse)
			if brokerResponse.Err != nil {
				if errors.Is(brokerResponse.Err, broker.KafkaError) {
					log.Error("broker error", "err", brokerResponse.Err)
					continue
				}
				log.Error("broker response error on message", "err", brokerResponse.Err)
				_, err := userStorage.UpdateSendStatus(context.Background(), brokerResponse.UserUUID, "failed")
				if err != nil {
					log.Error("failed to update message status", "err", err.Error())
				}
				continue
			}
			_, err := userStorage.UpdateSendStatus(context.Background(), brokerResponse.UserUUID, "successful")
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

// HealthCheck returns service health check
func (a *Auth) HealthCheck(ctx context.Context) error {
	log := a.log.With(
		slog.String("info", "SERVICE LAYER: metrics_service.HealthCheck"),
	)
	log.Info("starts getting health check")
	defer log.Info("finish getting health check")
	//TODO: add healthCheck
	//return a.healthChecker.HealthCheck(ctx)
	return nil
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
) (string, string, error) {
	ctx, span := tracer.Start(ctx, "service layer: login",
		trace.WithAttributes(attribute.String("handler", "login")))
	defer span.End()

	md, _ := metadata.FromIncomingContext(ctx)
	a.log.Info("span",
		"time", md.Get("timestamp"),
		"user-id", md.Get("user-id"),
		"x-trace-id", md.Get("x-trace-id"),
	)

	ctx, usrWithTokens, err := a.generateRefreshAccessToken(ctx, email)
	if err != nil {
		a.log.Error("Generation token failed:", err)
		return "", "", fmt.Errorf("generation token failed: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword(
		usrWithTokens.user.PassHash, []byte(password),
	); err != nil {
		a.log.Warn("invalid credentials")
		return "", "", fmt.Errorf("invalid credentials: %w", ErrInvalidCredentials)
	}
	return usrWithTokens.accessToken, usrWithTokens.refreshToken, nil
}

func (a *Auth) Refresh(
	ctx context.Context,
	token string,
) (string, string, error) {
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
	ctx, claims, err := a.validateToken(ctx, token)
	if err != nil {
		return "", "", ErrTokenRevoked
	}
	ttl := time.Duration(claims["exp"].(float64)-float64(time.Now().Unix())) * time.Second
	if err != nil {
		log.Info("failed validate token: ", "err", err.Error())
		return "", "", err
	}
	log.Info("validate token successfully")
	if claims["token_type"].(string) == "access" {
		return "", "", ErrTokenWrongType
	}
	email, ok := claims["email"].(string)
	if !ok {
		log.Error("token validation failed")
		return "", "", fmt.Errorf("token validation failed")
	}
	ctx, usrWithTokens, err := a.generateRefreshAccessToken(ctx, email)
	if err != nil {
		a.log.Error("failed to generate tokens", "err", err.Error())
		return "", "", err
	}
	a.log.Info("saving refresh token to redis")
	ctx, err = a.tokenStorage.SaveToken(ctx, token, ttl)
	if err != nil {
		a.log.Error("failed to save token", err.Error())
		return "", "", err
	}
	a.log.Info(" token saved to redis successfully")
	return usrWithTokens.accessToken, usrWithTokens.refreshToken, nil
}

func (a *Auth) Register(
	ctx context.Context,
	email string,
	password string,
) (string, error) {

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
		[]byte(password), bcrypt.DefaultCost,
	)
	if err != nil {
		log.Error("failed to generate password hash", err.Error())
		return "", fmt.Errorf("%s: %w", op, err)
	}
	ctx, id, err := a.userStorage.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", err.Error())
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registrated")

	RegirationMsg := registration_v1.RegistrationMessage{
		Email:    email,
		FullName: "Alex Black",
	}
	err = a.producer.Send(&RegirationMsg, "registration", id)
	if err != nil {
		// no return here with err!!!, we do continue working (so-called soft degradation)
		log.Error("Sending message to broker failed")
	}
	return id, nil
}

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
	ctx, user, err := a.userStorage.GetUser(ctx, userID)
	if err != nil {
		log.Error("failed to extract user", err.Error())
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user from database extracted")
	return user.IsUserAmin(), nil
}

func (a *Auth) Logout(
	ctx context.Context,
	token string,
) (success bool, err error) {

	log := a.log.With(
		slog.String("info", "SERVICE LAYER: auth_service.Logout"),
		slog.String("trace-id", "trace-id from opentelemetry"),
		slog.String("user-id", "user-id from opentelemetry extracted from jwt"),
	)

	log.Info("starting validate token")
	ctx, claims, err := a.validateToken(ctx, token)
	if err != nil {
		log.Info("failed validate token: ", err.Error())
		return false, err
	}
	ttl := time.Duration(claims["exp"].(float64)-float64(time.Now().Unix())) * time.Second

	log.Info("validate token successfully")
	log.Info("saving token to redis")

	ctx, err = a.tokenStorage.SaveToken(ctx, token, ttl)
	if err != nil {
		log.Error("failed to save token", err.Error())
		return false, err
	}
	log.Info("token saved to redis successfully")
	return true, nil
}

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
		return ctx, jwt.MapClaims{}, err
	}
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return ctx, jwt.MapClaims{}, ErrTokenParsing
	}
	// check ttl
	ttl := time.Duration(claims["exp"].(float64)-float64(time.Now().Unix())) * time.Second
	if ttl < 0 {
		return ctx, jwt.MapClaims{}, ErrTokenTtlExpired
	}
	// check type of token
	if (claims["token_type"] != "refresh") && claims["token_type"] != "access" {
		return ctx, jwt.MapClaims{}, ErrTokenWrongType
	}
	// check if token exists in redis

	ctx, value, err := a.tokenStorage.CheckTokenExists(ctx, token)
	if err != nil {
		return ctx, jwt.MapClaims{}, fmt.Errorf("validateToken: %w", err)
	}
	if value == TokenRevoked {
		return ctx, jwt.MapClaims{}, ErrTokenRevoked
	}
	return ctx, claims, nil
}

type userWithTokens struct {
	user         *models.User
	accessToken  string
	refreshToken string
	err          error
}

func (a *Auth) generateRefreshAccessToken(
	ctx context.Context,
	email string,
) (context.Context, userWithTokens, error) {

	ctx, user, err := a.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage2.ErrUserNotFound) {
			return ctx,
				userWithTokens{
					user:         nil,
					accessToken:  "",
					refreshToken: "",
				}, ErrInvalidCredentials
		}
		return ctx,
			userWithTokens{
				user:         nil,
				accessToken:  "",
				refreshToken: "",
			}, err
	}

	accessToken, err := jwtlib.NewToken(user, a.cfg, "access")
	if err != nil {
		return ctx,
			userWithTokens{
				user:         nil,
				accessToken:  "",
				refreshToken: "",
			}, fmt.Errorf("accessToken generation failed: %w", err)
	}
	refreshToken, err := jwtlib.NewToken(user, a.cfg, "refresh")
	if err != nil {
		return ctx,
			userWithTokens{
				user:         nil,
				accessToken:  "",
				refreshToken: "",
			}, fmt.Errorf("refreshToken generation failed: %w", err)
	}
	return ctx,
		userWithTokens{
			user:         &user,
			accessToken:  accessToken,
			refreshToken: refreshToken,
		}, nil
}
