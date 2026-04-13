CREATE TABLE IF NOT EXISTS instruments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL UNIQUE,
    provider_symbol TEXT NOT NULL,
    asset_type TEXT NOT NULL,
    exchange TEXT,
    quote_currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS portfolios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    base_currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    FOREIGN KEY (instrument_id) REFERENCES instruments(id),
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id),
    UNIQUE (portfolio_id, instrument_id)
);

CREATE INDEX IF NOT EXISTS idx_positions_portfolio_id
ON positions(portfolio_id);

CREATE INDEX IF NOT EXISTS idx_positions_instrument_id
ON positions(instrument_id);

CREATE TABLE IF NOT EXISTS price_cache (
    ticker TEXT PRIMARY KEY,
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
    fetched_at TIMESTAMP NOT NULL,
    PRIMARY KEY (base_currency, quote_currency)
);
