package patroni

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexBlackNn/authloyalty/loyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/domain"
	"github.com/AlexBlackNn/authloyalty/loyalty/internal/storage"
	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Storage struct {
	dbRead  *sql.DB
	dbWrite *sql.DB
}

const (
	Deposit           = "d"
	CheckViolationErr = "23514"
)

var tracer = otel.Tracer("loyalty service")

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
) (*domain.UserLoyalty, error) {
	ctx, span := tracer.Start(
		ctx, "data layer Patroni: GetLoyalty",
		trace.WithAttributes(attribute.String("handler", "GetLoyalty")),
	)
	defer span.End()

	query := "SELECT balance FROM loyalty_app.accounts WHERE uuid = $1;"
	err := s.dbRead.QueryRowContext(ctx, query, userLoyalty.UUID).Scan(&userLoyalty.Balance)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"DATA LAYER: storage.postgres.GetLoyalty: %w",
				storage.ErrUserNotFound,
			)
		}
		return nil, fmt.Errorf(
			"DATA LAYER: storage.postgres.GetLoyalty: %w",
			err,
		)
	}
	return userLoyalty, nil
}

func (s *Storage) AddLoyalty(
	ctx context.Context,
	userLoyalty *domain.UserLoyalty,
) (*domain.UserLoyalty, error) {
	ctx, span := tracer.Start(
		ctx, "data layer Patroni: AddLoyalty",
		trace.WithAttributes(attribute.String("handler", "AddLoyalty")),
	)
	defer span.End()

	//1. Open transaction
	tx, err := s.dbWrite.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	fmt.Println("0000000000000000000000")
	balance := userLoyalty.Balance

	//2. Block required row to avoid changing from other transactions
	var userLoyaltyBlocked domain.UserLoyalty
	query := "SELECT uuid, balance FROM loyalty_app.accounts WHERE uuid = $1 FOR UPDATE;"
	err = tx.QueryRowContext(ctx, query, userLoyalty.UUID).Scan(&userLoyaltyBlocked.UUID, &userLoyaltyBlocked.Balance)
	fmt.Println("0101010101010", err, userLoyalty, userLoyaltyBlocked)
	if err != nil {
		//3. If no row is selected
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("0101010101010", userLoyalty.Comment, userLoyalty.Operation)
			// 3.1 and operation is a registration then create new row in "accounts" and "loyalty_transactions" tables
			if userLoyalty.Operation == "registration" {
				query = "INSERT INTO loyalty_app.accounts (uuid, balance) VALUES ($1, $2) RETURNING uuid"
				err = tx.QueryRowContext(ctx, query, userLoyalty.UUID, userLoyalty.Balance).Scan(&userLoyalty.UUID)
				if err != nil {
					return nil, err
				}
				query = "INSERT INTO loyalty_app.loyalty_transactions (account_uuid, transaction_amount, transaction_type, comment) VALUES ($1, $2, $3, $4);"
				fmt.Println("11111111111111111111111111111111111111111111111111111111111", userLoyalty)
				_, err = tx.ExecContext(ctx, query, userLoyalty.UUID, userLoyalty.Balance, Deposit, userLoyalty.Comment)
				if err != nil {
					return nil, err
				}
				return userLoyalty, tx.Commit()
			}
		}
		// 3.2 and operation is NOT a registration (withdraw and deposit loyalty is forbidden if user is not registered)
		return nil, storage.ErrUserNotFound
	}

	// 4. if user account exists, try to update account
	if userLoyalty.Operation == "d" {
		// 4.1 if deposit
		query = "UPDATE loyalty_app.accounts SET balance = balance + $1 WHERE uuid = $2 RETURNING balance;"
	} else if userLoyalty.Operation == "w" {
		// 4.2 if withdraw
		query = "UPDATE loyalty_app.accounts SET balance = balance - $1 WHERE uuid = $2 RETURNING balance;"
	} else {
		// 4.3 if other type of operations (i.e."registration" - to create exactly only once "registration" operation)
		return nil, storage.ErrWrongParamType
	}
	err = tx.QueryRowContext(ctx, query, balance, userLoyalty.UUID).Scan(&userLoyalty.Balance)

	// https://www.postgresql.org/docs/16/errcodes-appendix.html
	var pgerr *pgconn.PgError

	if err != nil {
		if errors.As(err, &pgerr) {
			if pgerr.Code == CheckViolationErr {
				return nil, storage.ErrNegativeBalance
			}
		}
		return nil, storage.ErrInternalErr
	}

	// 5. Write data to account_transaction
	query = "INSERT INTO loyalty_app.loyalty_transactions (account_uuid, transaction_amount, transaction_type, comment) VALUES ($1, $2, $3, $4);"
	_, err = tx.ExecContext(ctx, query, userLoyalty.UUID, balance, userLoyalty.Operation, userLoyalty.Comment)
	if err != nil {
		return nil, err
	}
	return userLoyalty, tx.Commit()
}

func (s *Storage) HealthCheck(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "data layer Patroni: HealthCheck",
		trace.WithAttributes(attribute.String("handler", "HealthCheck")))
	defer span.End()
	// Pinger is an optional interface that may be implemented by a Conn. Then if driver
	// is changed need to be checked. https://pkg.go.dev/database/sql/driver#Pinger
	err := s.dbWrite.Ping()
	if err != nil {
		return fmt.Errorf(
			"DATA LAYER: storage.postgres.HealthCheck: couldn't ping databae  %w",
			err,
		)
	}
	return nil
}
