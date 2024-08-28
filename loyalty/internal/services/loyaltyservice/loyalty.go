package loyaltyservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type loyaltyStorage interface {
	AddLoyalty(
		ctx context.Context,
		loyalty *domain.UserLoyalty,
	) (context.Context, *domain.UserLoyalty, error)
	GetLoyalty(
		ctx context.Context,
		loyalty *domain.UserLoyalty,
	) (context.Context, *domain.UserLoyalty, error)
	HealthCheck(context.Context) (context.Context, error)
	Stop() error
}

type Loyalty struct {
	cfg          *config.Config
	log          *slog.Logger
	loyalStorage loyaltyStorage
}

// New returns a new instance of Auth service
func New(
	cfg *config.Config,
	log *slog.Logger,
	loyalStorage loyaltyStorage,
) *Loyalty {
	return &Loyalty{
		cfg:          cfg,
		log:          log,
		loyalStorage: loyalStorage,
	}
}

var tracer = otel.Tracer("loyalty service")

// HealthCheck returns service health check.
func (l *Loyalty) HealthCheck(ctx context.Context) (context.Context, error) {
	log := l.log.With(
		slog.String("info", "SERVICE LAYER: HealthCheck"),
	)
	log.Info("starts getting health check")
	defer log.Info("finish getting health check")
	return l.loyalStorage.HealthCheck(ctx)
}

func (l *Loyalty) GetLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (context.Context, *domain.UserLoyalty, error) {
	const op = "SERVICE LAYER: auth_service.RegisterNewUser"

	ctx, span := tracer.Start(ctx, "service layer: GetLoyalty",
		trace.WithAttributes(attribute.String("handler", "GetLoyalty")))
	defer span.End()

	log := l.log.With(
		slog.String("trace-id", "trace-id"),
		slog.String("user-id", "user-id"),
	)
	log.Info("getting loyalty for user")

	ctx, userLoyalty, err := l.loyalStorage.GetLoyalty(ctx, userLoyalty)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return ctx, nil, ErrUserNotFound
		}
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(fmt.Errorf("%s: failed to get loyalty: %w", op, err))
		log.Error("failed to get loyalty", "err", err.Error())
		return ctx, nil, fmt.Errorf("%s: %w", op, err)
	}
	span.AddEvent(
		"user loyalty extracted",
		trace.WithAttributes(
			attribute.String("user-id", userLoyalty.UUID),
			attribute.Int("user-id", userLoyalty.Balance),
		))

	return ctx, userLoyalty, nil
}

func (l *Loyalty) AddLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (context.Context, *domain.UserLoyalty, error) {
	const op = "SERVICE LAYER: auth_service.RegisterNewUser"

	ctx, span := tracer.Start(ctx, "service layer: AddLoyalty",
		trace.WithAttributes(attribute.String("handler", "AddLoyalty")))
	defer span.End()

	log := l.log.With(
		slog.String("trace-id", "trace-id"),
		slog.String("user-id", "user-id"),
	)
	log.Info("add loyalty to user")

	ctx, userLoyalty, err := l.loyalStorage.AddLoyalty(ctx, userLoyalty)
	if err != nil {
		if errors.Is(err, storage.ErrNegativeBalance) {
			return ctx, nil, ErrNegativeBalance
		}
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(fmt.Errorf("%s: failed to get loyalty: %w", op, err))
		log.Error("failed to get loyalty", "err", err.Error())
		return ctx, nil, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("------------------", userLoyalty)
	span.AddEvent(
		"user loyalty extracted",
		trace.WithAttributes(
			attribute.String("user-id", userLoyalty.UUID),
			attribute.Int("user-id", userLoyalty.Balance),
		))

	return ctx, userLoyalty, nil
}
