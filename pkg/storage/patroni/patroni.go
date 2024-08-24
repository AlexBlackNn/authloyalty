package patroni

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/pkg/storage"
	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// https://www.postgresql.org/docs/11/errcodes-appendix.html
const UniqueViolation = "23505"

type Storage struct {
	dbRead  *sql.DB
	dbWrite *sql.DB
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

// SaveUser saves user to db.
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (context.Context, string, error) {
	ctx, span := tracer.Start(
		ctx, "data layer Patroni: SaveUser",
		trace.WithAttributes(attribute.String("handler", "SaveUser")),
	)
	defer span.End()

	var uuid string
	query := "INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING uuid"
	err := s.dbWrite.QueryRowContext(ctx, query, email, passHash).Scan(&uuid)
	// https://www.postgresql.org/docs/11/protocol-error-fields.html
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		if pgerr.Code == UniqueViolation {
			return ctx, "", storage.ErrUserExists
		}
	}
	if err != nil {
		return ctx, "", fmt.Errorf(
			"DATA LAYER: storage.postgres.SaveUser: couldn't save user  %w",
			err,
		)
	}
	return ctx, uuid, nil
}

func (s *Storage) GetUser(ctx context.Context, uuid string) (context.Context, domain.User, error) {
	ctx, span := tracer.Start(ctx, "data layer Patroni: GetUser",
		trace.WithAttributes(attribute.String("handler", "GetUser")))
	defer span.End()

	query := "SELECT uuid, email, pass_hash, is_admin FROM users WHERE (uuid = $1);"
	row := s.dbRead.QueryRowContext(ctx, query, uuid)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PassHash, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ctx, domain.User{}, fmt.Errorf(
				"DATA LAYER: storage.postgres.GetUser: %w",
				storage.ErrUserNotFound,
			)
		}
		return ctx, domain.User{}, fmt.Errorf(
			"DATA LAYER: storage.postgres.GetUser: %w",
			err,
		)
	}
	return ctx, user, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (context.Context, domain.User, error) {
	ctx, span := tracer.Start(ctx, "data layer Patroni: GetUser",
		trace.WithAttributes(attribute.String("handler", "GetUser")))
	defer span.End()

	query := "SELECT uuid, email, pass_hash, is_admin FROM users WHERE (email = $1);"
	row := s.dbRead.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PassHash, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ctx, domain.User{}, fmt.Errorf(
				"DATA LAYER: storage.postgres.GetUser: %w",
				storage.ErrUserNotFound,
			)
		}
		return ctx, domain.User{}, fmt.Errorf(
			"DATA LAYER: storage.postgres.GetUser: %w",
			err,
		)
	}
	return ctx, user, nil
}

// UpdateSendStatus updates message send status.
func (s *Storage) UpdateSendStatus(ctx context.Context, uuid string, status string) (context.Context, error) {
	ctx, span := tracer.Start(ctx, "data layer Patroni: UpdateSendStatus",
		trace.WithAttributes(attribute.String("handler", "UpdateSendStatus")))
	defer span.End()

	query := "UPDATE users SET message_status=$2 WHERE uuid = $1;"
	_, err := s.dbWrite.ExecContext(ctx, query, uuid, status)
	if err != nil {
		return ctx, fmt.Errorf(
			"DATA LAYER: storage.postgres.UpdateSendStatus: couldn't update message registration status delivery  %w",
			err,
		)
	}
	return ctx, nil
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
