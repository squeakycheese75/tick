-- name: ListPositionsByPortfolio :many
SELECT
    p.instrument_id,
    pf.name,
    i.symbol,
    p.quantity,
    p.avg_cost,
    p.currency,
    i.asset_type,
    i.exchange,
    i.quote_currency
FROM positions AS p
JOIN portfolios AS pf ON p.portfolio_id = pf.id
JOIN instruments AS i ON p.instrument_id = i.id
WHERE p.portfolio_id = ?
ORDER BY i.symbol ASC;

-- name: CreatePosition :exec
INSERT INTO positions (
  instrument_id,
  portfolio_id,
  quantity,
  avg_cost,
  currency
) VALUES (
  ?, ?, ?, ?, ?
)
ON CONFLICT(portfolio_id, instrument_id) DO UPDATE SET
    quantity = excluded.quantity,
    avg_cost = excluded.avg_cost,
    currency = excluded.currency;
;

-- name: GetPortfolioByName :one
SELECT id, name, base_currency, created_at, updated_at
FROM portfolios
WHERE name = ?;

-- name: CreatePortfolio :exec
INSERT INTO portfolios (name, base_currency)
VALUES (?, ?)
ON CONFLICT(name) DO UPDATE SET
    base_currency = excluded.base_currency;

-- name: GetInstrumentBySymbolAndExchange :one
SELECT id, symbol, provider_symbol, asset_type, exchange, quote_currency, created_at, updated_at
FROM instruments
WHERE symbol = ? AND exchange = ?;

-- name: CreateInstrument :one
INSERT INTO instruments (
  symbol,
  provider_symbol,
  asset_type,
  exchange,
  quote_currency
) VALUES (
  ?, ?, ?, ?, ?
)
RETURNING id;

-- name: GetPriceCacheByTicker :one
SELECT
    ticker,
    price,
    price_currency,
    previous_close,
    change,
    change_percent,
    source,
    fetched_at
FROM price_cache
WHERE ticker = ?;

-- name: UpsertPriceCache :exec
INSERT INTO price_cache (
    ticker,
    price,
    price_currency,
    previous_close,
    change,
    change_percent,
    source,
    fetched_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(ticker) DO UPDATE SET
    price = excluded.price,
    price_currency = excluded.price_currency,
    previous_close = excluded.previous_close,
    change = excluded.change,
    change_percent = excluded.change_percent,
    source = excluded.source,
    fetched_at = excluded.fetched_at;

-- name: GetFXCacheByPair :one
SELECT
    base_currency,
    quote_currency,
    rate,
    source,
    fetched_at
FROM fx_cache
WHERE base_currency = ? AND quote_currency = ?;

-- name: UpsertFXCache :exec
INSERT INTO fx_cache (
    base_currency,
    quote_currency,
    rate,
    source,
    fetched_at
) VALUES (?, ?, ?, ?, ?)
ON CONFLICT(base_currency, quote_currency) DO UPDATE SET
    rate = excluded.rate,
    source = excluded.source,
    fetched_at = excluded.fetched_at;

-- name: CreatePortfolioSnapshot :one
INSERT INTO portfolio_snapshots (
    portfolio_name,
    base_currency,
    total_value,
    captured_at
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: CreatePortfolioSnapshotPosition :one
INSERT INTO portfolio_snapshot_positions (
    snapshot_id,
    symbol,
    quantity,
    instrument_currency,
    quoted_price,
    fx_rate,
    market_value_base,
    weight
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetLatestPortfolioSnapshot :one
SELECT *
FROM portfolio_snapshots
WHERE portfolio_name = ?
ORDER BY captured_at DESC, id DESC
LIMIT 1;

-- name: GetLatestPortfolioSnapshotBefore :one
SELECT *
FROM portfolio_snapshots
WHERE portfolio_name = ?
  AND captured_at < ?
ORDER BY captured_at DESC, id DESC
LIMIT 1;

-- name: ListPortfolioSnapshotPositionsBySnapshotID :many
SELECT *
FROM portfolio_snapshot_positions
WHERE snapshot_id = ?
ORDER BY symbol ASC;


-- name: CreateTarget :one
INSERT INTO portfolio_targets (
    portfolio_id,
    symbol,
    type,
    target_price,
    quote_currency
) VALUES (?, ?, ?, ?, ?)
RETURNING id;

-- name: ListTargetsByPortfolio :many
SELECT
    p.name,
    t.symbol,
    t.target_price,
    t.type,
    t.quote_currency,
    t.id
FROM portfolio_targets AS t
JOIN portfolios AS p ON t.portfolio_id = p.id
WHERE t.portfolio_id = ?
AND t.deleted_at IS NULL
ORDER BY t.symbol ASC;

-- name: DeleteTarget :execresult
DELETE FROM portfolio_targets
WHERE id = ?
AND portfolio_id = ?;
