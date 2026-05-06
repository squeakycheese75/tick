-- +goose Up
-- +goose StatementBegin
CREATE TABLE portfolio_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    type TEXT NOT NULL,
    target_price REAL NOT NULL,
    quote_currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    CONSTRAINT portfolio_targets_type_check
        CHECK (type IN ('take-profit', 'stop-loss'))
);

CREATE UNIQUE INDEX idx_portfolio_targets_unique_active
    ON portfolio_targets (portfolio_id, symbol, type)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_portfolio_targets_portfolio_id
    ON portfolio_targets (portfolio_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_portfolio_targets_portfolio_id;
DROP INDEX IF EXISTS idx_portfolio_targets_unique_active;
DROP TABLE IF EXISTS portfolio_targets;
-- +goose StatementEnd
