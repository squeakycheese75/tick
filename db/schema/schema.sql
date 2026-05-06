CREATE TABLE IF NOT EXISTS instruments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL UNIQUE,
    asset_type TEXT NOT NULL,
    exchange TEXT,
    quote_currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
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

CREATE TABLE IF NOT EXISTS price_cache (
    symbol TEXT PRIMARY KEY,
    provider_symbol TEXT,
    price REAL NOT NULL,
    price_currency TEXT NOT NULL,
    previous_close REAL NOT NULL,
    change REAL NOT NULL,
    change_percent REAL NOT NULL,
    source TEXT NOT NULL,
    fetched_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS fx_cache (
    base_currency TEXT NOT NULL,
    quote_currency TEXT NOT NULL,
    rate REAL NOT NULL,
    source TEXT NOT NULL,
    fetched_at TIMESTAMP NOT NULL,
    PRIMARY KEY (base_currency, quote_currency)
);

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
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (snapshot_id) REFERENCES portfolio_snapshots(id) ON DELETE CASCADE
);

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
