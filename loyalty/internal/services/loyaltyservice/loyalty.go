package loyaltyservice

import (
	"context"
	"log/slog"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"go.opentelemetry.io/otel"
)

type loyaltyStorage interface {
	AddLoyalty(
		ctx context.Context,
		loyalty domain.UserLoyalty,
	) (context.Context, domain.UserLoyalty, error)
	GetLoyalty(
		ctx context.Context,
		loyalty domain.UserLoyalty,
	) (context.Context, domain.UserLoyalty, error)
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

var tracer = otel.Tracer("sso service")

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
) (domain.UserLoyalty, error) {
	return domain.UserLoyalty{UUID: "e1d07926-0dda-4b1c-a284-1919da8da752", Value: 100}, nil
}

func (l *Loyalty) AddLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (domain.UserLoyalty, error) {
	return domain.UserLoyalty{UUID: "e1d07926-0dda-4b1c-a284-1919da8da752", Value: 100}, nil
}

func (l *Loyalty) SubLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (domain.UserLoyalty, error) {
	return domain.UserLoyalty{UUID: "e1d07926-0dda-4b1c-a284-1919da8da752", Value: 100}, nil
}
