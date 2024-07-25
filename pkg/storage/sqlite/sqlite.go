package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/domain/models"
	"github.com/AlexBlackNn/authloyalty/pkg/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "DATA LAYER: storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}

// SaveUser saves user to db.
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "DATA LAYER: storage.sqlite.SaveUser"

	query := "INSERT INTO users(email, pass_hash) VALUES(?, ?)"
	res, err := s.db.ExecContext(ctx, query, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUser(ctx context.Context, value any) (models.User, error) {
	const op = "DATA LAYER: storage.sqlite.User"
	var row *sql.Row
	query := "SELECT id, email, pass_hash, is_admin FROM users WHERE (email = ? AND ? IS NOT NULL) OR (id = ? AND ? IS NOT NULL);"
	switch sqlParam := value.(type) {
	case int:
		row = s.db.QueryRowContext(ctx, query, nil, nil, sqlParam, sqlParam)
	case string:
		row = s.db.QueryRowContext(ctx, query, sqlParam, sqlParam, nil, nil)
	default:
		return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrWrongParamType)
	}

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.PassHash, &user.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

// App returns app by id.
func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	const op = "DATA LAYER: storage.sqlite.App"
	query := "SELECT id, name, secret FROM apps WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)
	var app models.App
	err := row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}
