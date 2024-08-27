package patroni

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/pkg/storage"
	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Storage struct {
	dbRead  *sql.DB
	dbWrite *sql.DB
}

type loyaltyStorage interface {
	AddLoyaly(
		ctx context.Context,
		loyalty domain.UserLoyalty,
	) (context.Context, domain.UserLoyalty, error)
	SubLoyalty(
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

var tracer = otel.Tracer("sso service")

func New(cfg *config.Config) (*Storage, error) {
	dbWrite, err := otelsql.Open("pgx", cfg.StoragePatroni.Master)
	if err != nil {
		return nil, fmt.Errorf(
			"DATA LAYER: storage.postgres.New: couldn't open a database for Write: %w",
			storage.ErrConnection,
		)
	}
	dbRead, err := otelsql.Open("pgx", cfg.StoragePatroni.Slave)
	if err != nil {
		return nil, fmt.Errorf(
			"DATA LAYER: storage.postgres.New: couldn't open a database for Read: %w",
			storage.ErrConnection,
		)
	}
	// Open may just validate its arguments without creating a connection to the database.
	// To verify that the data source name is valid, call DB.Ping.
	err = dbRead.Ping()
	if err != nil {
		return nil, fmt.Errorf(
			"DATA LAYER: storage.postgres.New: couldn't connect to database for Read: %w", err,
		)
	}
	err = dbWrite.Ping()
	if err != nil {
		return nil, fmt.Errorf(
			"DATA LAYER: storage.postgres.New: couldn't connect to database for Write: %w", err,
		)
	}
	return &Storage{dbRead: dbRead, dbWrite: dbWrite}, nil
}

func (s *Storage) Stop() error {
	var err1, err2 error
	if s.dbRead != nil {
		err1 = s.dbWrite.Close()
	}
	if s.dbWrite != nil {
		err2 = s.dbRead.Close()
	}
	return errors.Join(err1, err2)
}

func (s *Storage) GetLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (context.Context, *domain.UserLoyalty, error) {
	ctx, span := tracer.Start(
		ctx, "data layer Patroni: SaveUser",
		trace.WithAttributes(attribute.String("handler", "SaveUser")),
	)
	defer span.End()
	return ctx, userLoyalty, nil
}

func (s *Storage) AddLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (context.Context, *domain.UserLoyalty, error) {
	ctx, span := tracer.Start(
		ctx, "data layer Patroni: SaveUser",
		trace.WithAttributes(attribute.String("handler", "SaveUser")),
	)
	defer span.End()
	return ctx, userLoyalty, nil
}

func (s *Storage) HealthCheck(ctx context.Context) (context.Context, error) {
	ctx, span := tracer.Start(ctx, "data layer Patroni: HealthCheck",
		trace.WithAttributes(attribute.String("handler", "HealthCheck")))
	defer span.End()
	// Pinger is an optional interface that may be implemented by a Conn. Then if driver
	// is changed need to be checked. https://pkg.go.dev/database/sql/driver#Pinger
	err := s.dbWrite.Ping()
	if err != nil {
		return ctx, fmt.Errorf(
			"DATA LAYER: storage.postgres.HealthCheck: couldn't ping databae  %w",
			err,
		)
	}
	return ctx, nil
}
