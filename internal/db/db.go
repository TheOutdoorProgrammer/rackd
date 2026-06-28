// Package db opens the SQLite database and runs migrations. All application data
// is encrypted by the vault before it reaches these tables; this package is only
// concerned with storage, pragmas, and schema.
package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // pure-Go driver, registered as "sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens (creating if needed) the SQLite database at path, applies pragmas
// suited to a single-writer embedded app, and runs all pending migrations.
func Open(path string) (*sql.DB, error) {
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)&_pragma=busy_timeout(5000)"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open sqlite: %w", err)
	}
	// modernc/sqlite is happiest with a single connection; serialize writers.
	database.SetMaxOpenConns(1)
	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	if err := migrate(database); err != nil {
		_ = database.Close()
		return nil, err
	}
	return database, nil
}

func migrate(database *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("db: goose dialect: %w", err)
	}
	if err := goose.Up(database, "migrations"); err != nil {
		return fmt.Errorf("db: run migrations: %w", err)
	}
	return nil
}
