package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	"github.com/AlexBlackNn/authloyalty/pkg/storage"
	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const ErrCodeUserAlreadyExists = "23505"

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
			err,
		)
	}
	dbRead, err := otelsql.Open("pgx", cfg.StoragePatroni.Slave)
	if err != nil {
		return nil, fmt.Errorf(
			"DATA LAYER: storage.postgres.New: couldn't open a database for Read: %w",
			err,
		)
	}
	return &Storage{dbRead: dbRead, dbWrite: dbWrite}, nil
}

func (s *Storage) Stop() error {
	err1 := s.dbWrite.Close()
	err2 := s.dbRead.Close()
	return fmt.Errorf("%w, %w", err1, err2)
}

// SaveUser saves user to db.
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (context.Context, string, error) {
	ctx, span := tracer.Start(ctx, "data layer Patroni: SaveUser",
		trace.WithAttributes(attribute.String("handler", "SaveUser")))
	defer span.End()

	var id string
	query := "INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING uuid"
	err := s.dbWrite.QueryRowContext(ctx, query, email, passHash).Scan(&id)
	// https://stackoverflow.com/questions/34963064/go-pq-and-postgres-appropriate-error-handling-for-constraints
	if err, ok := err.(*pgconn.PgError); ok {
		if err.Code == ErrCodeUserAlreadyExists {
			return ctx, "", storage.ErrUserExists
		}
	}
	if err != nil {
		return ctx, "", fmt.Errorf(
			"DATA LAYER: storage.postgres.SaveUser: couldn't save user  %w",
			err,
		)
	}

	return ctx, id, nil
}

func (s *Storage) GetUser(ctx context.Context, value any) (context.Context, models.User, error) {
	ctx, span := tracer.Start(ctx, "data layer Patroni: GetUser",
		trace.WithAttributes(attribute.String("handler", "GetUser")))
	defer span.End()

	var row *sql.Row
	switch sqlParam := value.(type) {
	case int:
		query := "SELECT uuid, email, pass_hash, is_admin FROM users WHERE (uuid = $1);"
		row = s.dbRead.QueryRowContext(ctx, query, sqlParam)
	case string:
		query := "SELECT uuid, email, pass_hash, is_admin FROM users WHERE (email = $1);"
		row = s.dbRead.QueryRowContext(ctx, query, sqlParam)
	default:
		return ctx, models.User{}, fmt.Errorf(
			"DATA LAYER: storage.postgres.GetUser: %w",
			storage.ErrWrongParamType,
		)
	}

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PassHash, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ctx, models.User{}, fmt.Errorf(
				"DATA LAYER: storage.postgres.GetUser: %w",
				storage.ErrUserNotFound,
			)
		}
		return ctx, models.User{}, fmt.Errorf(
			"DATA LAYER: storage.postgres.GetUser: %w",
			err,
		)
	}
	return ctx, user, nil
}

// UpdateSendStatus updates message send status.
func (s *Storage) UpdateSendStatus(ctx context.Context, uuid string) (context.Context, error) {
	ctx, span := tracer.Start(ctx, "data layer Patroni: UpdateSendStatus",
		trace.WithAttributes(attribute.String("handler", "UpdateSendStatus")))
	defer span.End()

	query := "UPDATE users SET message_status='successful' WHERE uuid = $1;"
	_, err := s.dbWrite.ExecContext(ctx, query, uuid)
	if err != nil {
		return ctx, fmt.Errorf(
			"DATA LAYER: storage.postgres.UpdateSendStatus: couldn't update message registration status deluvery  %w",
			err,
		)
	}
	return ctx, nil
}
