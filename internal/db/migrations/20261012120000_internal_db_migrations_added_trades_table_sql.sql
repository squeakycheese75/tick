-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS trades (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id  INTEGER NOT NULL,
    instrument_id INTEGER NOT NULL,
    trade_key     TEXT NOT NULL,
    side          TEXT NOT NULL,
    quantity      REAL NOT NULL,
    price         REAL NOT NULL,
    currency      TEXT NOT NULL,
    fee           REAL NOT NULL DEFAULT 0,
    fx_rate       REAL NOT NULL DEFAULT 1,
    traded_at     TIMESTAMP NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id),
    FOREIGN KEY (instrument_id) REFERENCES instruments(id),
    UNIQUE (portfolio_id, trade_key)
);

CREATE INDEX IF NOT EXISTS idx_trades_portfolio_traded_at
ON trades(portfolio_id, traded_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trades;
-- +goose StatementEnd
