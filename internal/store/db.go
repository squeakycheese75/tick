package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	sqlDB *sql.DB
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	wrapped := &DB{sqlDB: db}

	if err := wrapped.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return wrapped, nil
}

func (db *DB) Close() error {
	return db.sqlDB.Close()
}

func (db *DB) migrate(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS positions (
    portfolio_name TEXT NOT NULL,
    ticker TEXT NOT NULL,
    quantity REAL NOT NULL,
    avg_cost REAL NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    PRIMARY KEY (portfolio_name, ticker)
);
`

	if _, err := db.sqlDB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("migrate positions table: %w", err)
	}

	const query2 = `
CREATE TABLE IF NOT EXISTS portfolios (
    name TEXT PRIMARY KEY,
    base_currency TEXT NOT NULL
);
`

	if _, err := db.sqlDB.ExecContext(ctx, query2); err != nil {
		return fmt.Errorf("migrate positions table: %w", err)
	}

	return nil
}
