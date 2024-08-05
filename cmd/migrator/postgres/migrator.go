package main

import (
	"errors"
	"flag"
	"fmt"
	// migration lib
	"github.com/golang-migrate/migrate/v4"
	// driver for migration applying postgres
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// driver to get migrations from files (*.sql in our case)
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath, migrationsTable string

	flag.StringVar(
		&migrationsPath,
		"migrations-path",
		"",
		"path to migrations",
	)
	flag.StringVar(
		&migrationsTable,
		"migrations-table",
		"migrations",
		"name of migration table, where migrator writes own data",
	)
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf(
			"postgres://postgres:postgres@localhost:5000/postgres?sslmode=disable",
		),
	)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			// TODO: change to logger
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
	fmt.Println("migrations applied successfully")
}
