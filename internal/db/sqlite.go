package db

import (
	"database/sql"

	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

type DB struct {
	SqlDB *sql.DB
}

func OpenAndMigrateSqlite(dsn string) (*DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	wrapped := &DB{SqlDB: db}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	if err := goose.Up(db, "internal/db/migrations"); err != nil {
		_ = db.Close()
		return nil, err
	}

	return wrapped, nil
}

func (db *DB) Close() error {
	return db.SqlDB.Close()
}
