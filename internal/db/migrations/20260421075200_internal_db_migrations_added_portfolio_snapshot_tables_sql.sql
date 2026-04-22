-- +goose Up
-- +goose StatementBegin
CREATE TABLE portfolio_snapshots (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_name TEXT NOT NULL,
    base_currency  TEXT NOT NULL,
    total_value    REAL NOT NULL,
    captured_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE portfolio_snapshot_positions (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    snapshot_id         INTEGER NOT NULL,
    symbol              TEXT NOT NULL,
    quantity            REAL NOT NULL,
    instrument_currency TEXT NOT NULL,
    quoted_price        REAL NOT NULL,
    fx_rate             REAL NOT NULL,
    market_value_base   REAL NOT NULL,
    weight              REAL NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (snapshot_id) REFERENCES portfolio_snapshots(id) ON DELETE CASCADE
);

CREATE INDEX idx_portfolio_snapshots_portfolio_captured_at
    ON portfolio_snapshots (portfolio_name, captured_at DESC);

CREATE INDEX idx_portfolio_snapshot_positions_snapshot_id
    ON portfolio_snapshot_positions (snapshot_id);

CREATE INDEX idx_portfolio_snapshot_positions_snapshot_symbol
    ON portfolio_snapshot_positions (snapshot_id, symbol);
-- +goose StatementEnd
    
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_portfolio_snapshot_positions_snapshot_symbol;
DROP INDEX IF EXISTS idx_portfolio_snapshot_positions_snapshot_id;
DROP INDEX IF EXISTS idx_portfolio_snapshots_portfolio_captured_at;
DROP TABLE IF EXISTS portfolio_snapshot_positions;
DROP TABLE IF EXISTS portfolio_snapshots;
-- +goose StatementEnd
