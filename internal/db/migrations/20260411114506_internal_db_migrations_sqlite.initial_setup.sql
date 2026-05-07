-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS instruments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    asset_type TEXT NOT NULL,
    exchange TEXT,
    quote_currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(symbol, exchange)
);

CREATE TABLE IF NOT EXISTS portfolios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    base_currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    instrument_id INTEGER NOT NULL,
    portfolio_id INTEGER NOT NULL,
    quantity REAL NOT NULL,
    avg_cost REAL NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (instrument_id) REFERENCES instruments(id),
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id),
    UNIQUE (portfolio_id, instrument_id)
);

CREATE INDEX IF NOT EXISTS idx_positions_portfolio_id
ON positions(portfolio_id);

CREATE INDEX IF NOT EXISTS idx_positions_instrument_id
ON positions(instrument_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_positions_instrument_id;
DROP INDEX IF EXISTS idx_positions_portfolio_id;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS portfolios;
DROP TABLE IF EXISTS positions;
-- +goose StatementEnd
