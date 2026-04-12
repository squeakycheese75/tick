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
