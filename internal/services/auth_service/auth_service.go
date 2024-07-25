package auth_service

//SERVICE LAYER
// Auth(service layer) encapsulates userStorage (data layer)

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	jwtlib "github.com/AlexBlackNn/authloyalty/internal/lib/jwt"
	"github.com/AlexBlackNn/authloyalty/storage"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"time"
)

type Auth struct {
	log *slog.Logger
	// data layer
	userStorage storage.UserStorage
	// data layer
	tokenStorage storage.TokenStorage
	cfg          *config.Config
}

// New returns a new instance of Auth service
func New(
	log *slog.Logger,
// data layer
	userStorage storage.UserStorage,
// data layer
	tokenStorage storage.TokenStorage,

	cfg *config.Config,
) *Auth {
	return &Auth{
		log:          log,
		userStorage:  userStorage,
		tokenStorage: tokenStorage,
		cfg:          cfg,
	}
}

const (
	TokenRevoked = 1
)

var tracer = otel.Tracer("sso service")

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
) (string, string, error) {
	ctx, span := tracer.Start(ctx, "service layer: login",
		trace.WithAttributes(attribute.String("handler", "login")))
	defer span.End()

	md, _ := metadata.FromIncomingContext(ctx)
	a.log.Info("time: %v, userId: %v", md.Get("timestamp"), md.Get("user-id"))

	ctx, usrWithTokens, err := a.generateRefreshAccessToken(ctx, email)
	if err != nil {
		a.log.Error("Generation token failed:", err)
		return "", "", fmt.Errorf(
			"generation token failed: %w", err,
		)
	}

	if err := bcrypt.CompareHashAndPassword(
		usrWithTokens.user.PassHash, []byte(password),
	); err != nil {
		a.log.Info("invalid credentials")
		return "", "", fmt.Errorf(
			"invalid credentials: %w", ErrInvalidCredentials,
		)
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
		log.Info("failed validate token: ", err.Error())
		return "", "", err
	}
	log.Info("validate token successfully")
	if claims["token_type"].(string) == "access" {
		return "", "", ErrTokenWrongType
	}
	userID := int(claims["uid"].(float64))
	ctx, usrWithTokens, err := a.generateRefreshAccessToken(ctx, userID)
	if err != nil {
		a.log.Error("failed to generate tokens", err.Error())
		return "", "", err
	}
	a.log.Info("saving refresh token to redis")
	ctx, err = a.tokenStorage.SaveToken(ctx, token, ttl)
	if err != nil {
		a.log.Error("failed to save token", err.Error())
		return "", "", err
	}
	fmt.Println("888888888")
	a.log.Info(" token saved to redis successfully")
	return usrWithTokens.accessToken, usrWithTokens.refreshToken, nil
}

func (a *Auth) Register(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {

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
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	ctx, id, err := a.userStorage.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", err.Error())
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registrated")
	//TODO: add message to kafka

	return id, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int,
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
	value any,
) (context.Context, userWithTokens, error) {

	ctx, user, err := a.userStorage.GetUser(ctx, value)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
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
