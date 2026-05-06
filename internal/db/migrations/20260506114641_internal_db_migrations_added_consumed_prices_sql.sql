-- +goose Up
-- +goose StatementBegin
CREATE TABLE consumed_prices (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol          TEXT NOT NULL,
    source          TEXT NOT NULL,
    price           REAL NOT NULL,
    currency        TEXT NOT NULL,
    as_of           TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(symbol, source, as_of)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS consumed_prices;
-- +goose StatementEnd
