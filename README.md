# tick

Terminal-native portfolio and market intelligence tool.

`tick` is a Bloomberg-style CLI for developers and investors.\
It provides real-time portfolio valuation, risk insights, and market
context --- directly from your terminal.

------------------------------------------------------------------------

## Example Output

![tick daily output](docs/screenshot.png)

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

### Daily Brief (`tick daily`)

-   Portfolio overview
-   Top holdings
-   Risk summary
-   News per holding
-   Attention signals
-   Daily price moves (% change with arrows)

------------------------------------------------------------------------

## Getting Started

### Run locally

``` bash
go run ./cmd/tick daily
```

### Build the CLI

``` bash
go build -o bin/tick ./cmd/tick
./bin/tick daily
```

### Install globally (Go)

``` bash
go install github.com/squeakycheese75/tick/cmd/tick@latest
```

------------------------------------------------------------------------

## Example Usage

Create a portfolio:

``` bash
tick portfolio create main --base-currency EUR
```

Add positions:

``` bash
tick portfolio add NVDA --qty 10 --avg-cost 400 --currency USD --portfolio main
tick portfolio add ASML --qty 5 --avg-cost 850 --currency EUR --portfolio main
```

View summary:

``` bash
tick portfolio summary
```

Run daily brief:

``` bash
tick daily
```

------------------------------------------------------------------------

## Configuration

`tick` is configured via environment variables (supports `.env`).

### Example `.env`

``` env
PRICE_PROVIDER=finnhub
FX_PROVIDER=frankfurter

FINNHUB_API_KEY=your_api_key_here

CACHE_ENABLED=true
CACHE_PRICE_TTL=15m
CACHE_FX_TTL=12h
```

------------------------------------------------------------------------

## License

MIT License
