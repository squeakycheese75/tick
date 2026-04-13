package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"

	"github.com/pressly/goose/v3"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type DB struct {
	SqlDB *sql.DB
}

func OpenAndMigrateSqlite(dsn string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	wrapped := &DB{SqlDB: sqlDB}

	if err := goose.SetDialect("sqlite3"); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	// silent migrations
	originalWriter := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalWriter)

	// use the embedded migrations files

	migrationsFS, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("sub migrations fs: %w", err)
	}

	goose.SetBaseFS(migrationsFS)

	if err := goose.Up(sqlDB, "."); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return wrapped, nil
}

func (db *DB) Close() error {
	return db.SqlDB.Close()
}

func IsUniqueViolation(err error) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	switch sqliteErr.Code() {
	case sqlite3.SQLITE_CONSTRAINT_UNIQUE, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY:
		return true
	default:
		return false
	}
}
