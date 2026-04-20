# tick

Terminal-native portfolio tracker with live pricing, FX conversion, and instant PnL — built for developers.

---

## Quick Demo

```bash
tick portfolio create main --base-currency EUR

tick add NVDA --qty 10 --avg-cost 400
tick add MSTR --qty 10 --avg-cost 100

tick portfolio summary
```

Example output:

```
Portfolio: main

Base currency: EUR
Total value: 3127.53 EUR
Total cost: 4416.93 EUR
Total PnL: -1289.40 EUR (-29.19%)

Positions:
TICKER   QTY         PRICE         VALUE         COST          PNL           PNL %
NVDA     10.0000     201.68 USD    1713.09 EUR   3397.64 EUR   -1684.55 EUR  -49.58%
MSTR     10.0000     166.52 USD    1414.44 EUR   1019.29 EUR    395.15 EUR   38.77%
```

---

## Features

### Portfolio Management

- Create and manage portfolios
- Add positions with simple commands (`tick add NVDA`)
- Automatic instrument resolution (asset type, exchange, currency)
- Import portfolios from JSON
- Multi-currency support

### Valuation & Pricing

- Live price data (via Finnhub)
- FX conversion (via Frankfurter)
- Portfolio base currency normalization
- Cached pricing and FX

### Analysis

- Portfolio valuation and weights
- Cost basis and PnL tracking
- Per-position metrics:
  - market value
  - cost basis
  - unrealized PnL

### AI Analysis (`--ai`)

- Local LLM support via Ollama
- Generates concise insights
- Fully private

---

## Installation

```bash
brew tap squeakycheese75/tick
brew install tick
```

---

## Usage

### Create a portfolio

```bash
tick portfolio create main --base-currency EUR
```

### Add a position

```bash
tick add NVDA --qty 10 --avg-cost 400
```

> Instrument metadata is resolved automatically for supported symbols.

### Import a portfolio

```bash
tick portfolio import --file ./portfolio.json
```

---

## License

MIT
