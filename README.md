# tick

Terminal-native portfolio and market intelligence tool.

`tick` is a Bloomberg-style CLI for developers and investors.\
It provides real-time portfolio valuation, risk insights, and market
context --- directly from your terminal.

------------------------------------------------------------------------

## Features

### Portfolio Management

-   Create and manage portfolios
-   Add/update positions
-   Multi-currency support per position

### Valuation & Pricing

-   Live price data (via Finnhub)
-   FX conversion (via Frankfurter)
-   Portfolio base currency normalization
-   Cached pricing and FX (configurable TTLs)

### Analysis

-   Portfolio summary (weights, values)
-   Concentration risk analysis
-   Top holdings breakdown

### Daily Brief (`tick daily`)

-   Portfolio overview
-   Top holdings
-   Risk summary
-   News per holding
-   Attention signals
-   Daily price moves (% change with arrows)

### CLI Experience

-   Fast, terminal-first UX
-   Clean, readable output
-   Designed for daily usage

------------------------------------------------------------------------

## Getting Started

### Run locally

go run ./cmd/tick daily

### Build the CLI

go build -o bin/tick ./cmd/tick ./bin/tick daily

### Install globally (Go)

go install github.com/squeakycheese75/tick/cmd/tick@latest

------------------------------------------------------------------------

## Example Usage

Create a portfolio:

tick portfolio create main --base-currency EUR

Add positions:

tick portfolio add NVDA --qty 10 --avg-cost 400 --currency USD
--portfolio main tick portfolio add ASML --qty 5 --avg-cost 850
--currency EUR --portfolio main

View summary:

tick portfolio summary

Run daily brief:

tick daily

------------------------------------------------------------------------

## Configuration

`tick` is configured via environment variables (supports `.env`).

### Example `.env`

PRICE_PROVIDER=finnhub FX_PROVIDER=frankfurter

FINNHUB_API_KEY=your_api_key_here

CACHE_ENABLED=true CACHE_PRICE_TTL=15m CACHE_FX_TTL=12h

### Providers

-   Price: static, finnhub
-   FX: static, frankfurter

------------------------------------------------------------------------

## Project Structure

cmd/tick/ \# CLI entrypoint internal/ app/ \# wiring, config, provider
factories domain/ \# core models usecase/ \# application logic service/
\# reusable services adapters/market/ \# price + FX providers cli/ \#
rendering/output

------------------------------------------------------------------------

## Design Principles

-   Terminal-first
-   Local-first
-   Deterministic core
-   Composable architecture
-   Extensible
-   Grounded intelligence

------------------------------------------------------------------------

## Roadmap

### Near Term

-   Live price integration
-   FX conversion
-   Daily brief
-   News integration
-   Caching

### Next

-   AI-assisted portfolio analysis (local LLM)
-   Better instrument metadata
-   Improved terminal formatting
-   Historical performance tracking

### Future

-   Strategy simulation
-   Alerts and signals
-   Plugin/provider ecosystem

------------------------------------------------------------------------

## Versioning

Semantic versioning

Current stage: v0.x

------------------------------------------------------------------------

## License

MIT License
