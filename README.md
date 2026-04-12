# tick

Terminal-native portfolio and market intelligence tool.

`tick` is a Bloomberg-style CLI for developers and investors. It
provides portfolio valuation, risk insights, and market context — directly
from your terminal, with optional **local AI analysis**.

------------------------------------------------------------------------

## Features

### Portfolio Management

- Create and manage portfolios
- Add positions with full instrument metadata
- Import portfolios from JSON
- Multi-currency support per position

### Valuation & Pricing

- Live price data (via Finnhub)
- FX conversion (via Frankfurter)
- Portfolio base currency normalization
- Cached pricing and FX (configurable TTLs)

### Analysis

- Portfolio valuation and weights
- Concentration insights
- Per-position metrics:
  - market value
  - weights
  - FX-adjusted pricing
  - change in base currency

### AI Analysis (`--ai`)

- Local LLM support via Ollama
- Generates concise insights
- Fully private (no external API required)

------------------------------------------------------------------------

## Getting Started

### Run locally

```bash
go run ./cmd/tick daily
```

### With AI

```bash
go run ./cmd/tick daily --ai
```

### Build the CLI

```bash
go build -o bin/tick ./cmd/tick
./bin/tick daily
```

------------------------------------------------------------------------

## Example Usage

### Create a portfolio

```bash
tick portfolio create main --base-currency EUR
```

### Add a position

```bash
tick portfolio add NVDA \
  --portfolio main \
  --asset-type equity \
  --exchange NASDAQ \
  --quote-currency USD \
  --qty 10 \
  --avg-cost 400
```

> Note: Instruments are automatically created if they do not exist.

---

### Import a portfolio (recommended for development)

```bash
tick portfolio import --file ./testdata/portfolio-main.json
```

Example file:

```json
{
  "portfolioName": "main",
  "baseCurrency": "EUR",
  "positions": [
    {
      "symbol": "NVDA",
      "assetType": "equity",
      "exchange": "NASDAQ",
      "quoteCurrency": "USD",
      "quantity": 10,
      "avgCost": 400
    }
  ]
}
```

---

### Run daily brief

```bash
tick daily
```

### With AI

```bash
tick daily --ai
```

------------------------------------------------------------------------

## Configuration

`tick` is configured via environment variables (supports `.env`).

### Pricing & FX

```env
# Price data provider: static | finnhub
PRICE_PROVIDER=finnhub

# FX provider: static | frankfurter
FX_PROVIDER=frankfurter

# Required if using Finnhub
FINNHUB_API_KEY=your_api_key_here
```

---

### Caching

```env
CACHE_ENABLED=true
CACHE_PRICE_TTL=15m
CACHE_FX_TTL=12h
```

---

### LLM (Ollama)

```env
LLM_ENABLED=true
LLM_PROVIDER=ollama
LLM_BASE_URL=http://localhost:11434

# Model tag optional (":latest" assumed if omitted)
LLM_MODEL=llama3
```

Install and run Ollama:

```bash
brew install ollama
ollama pull llama3
ollama run llama3
```

---

## Architecture

- **domain** → core business entities
- **repository** → persistence layer (sqlc)
- **usecase** → application logic
- **analysis** → portfolio analytics engine
- **adapters** → external integrations (pricing, FX, LLM)
- **cmd** → CLI entrypoints (Cobra)

------------------------------------------------------------------------

## Development Notes

- Use `tick portfolio import` to quickly seed data
- Instruments are created automatically on first use
- Model names are normalized (`llama3` → `llama3:latest`)
- Ollama health is checked via `/api/version`

------------------------------------------------------------------------

## License

MIT License
