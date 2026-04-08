# tick

Terminal-native portfolio and market intelligence tool.

`tick` is a Bloomberg-style CLI for developers and investors.  
It helps you track your portfolio, understand exposure, and analyse assets — directly from your terminal.

---

## Features

- Portfolio tracking (multi-currency aware)
- Position management
- Portfolio summaries and weighting
- Currency-normalised valuation (base currency per portfolio)
- Clean CLI interface

---

## Getting Started

### Run locally

```bash
go run ./cmd/tick portfolio summary
```

### Build the CLI

```bash
go build -o bin/tick ./cmd/tick
./bin/tick portfolio summary
```

### Install globally (Go)

```bash
go install github.com/squeakycheese75/tick/cmd/tick@latest
```

Then:

```bash
tick portfolio summary
```

---

## Example Usage

Create a portfolio:

```bash
tick portfolio create main --base-currency EUR
```

Add positions:

```bash
tick portfolio add NVDA --qty 10 --avg-cost 400 --currency USD --portfolio main
tick portfolio add ASML --qty 5 --avg-cost 850 --currency EUR --portfolio main
```

View your portfolio:

```bash
tick portfolio summary
```

---

## Project Structure

```
cmd/tick/          # CLI entrypoint
internal/
  app/             # application services (use cases)
  domain/          # core models and logic
  adapters/        # DB, FX, pricing providers
  cli/             # rendering/output
```

---

## Design Principles

- **Terminal-first**: fast, minimal, composable
- **Local-first**: SQLite-backed, no cloud dependency
- **Deterministic core**: calculations handled in code
- **Modular architecture**: CLI is thin, logic lives in services
- **Extensible**: designed for future data sources and interfaces

---

## Roadmap

- [ ] Portfolio risk analysis
- [ ] Asset info and fundamentals
- [ ] News integration
- [ ] Synthetic portfolios and strategy simulation
- [ ] Better terminal output (tables, formatting)
- [ ] Live price and FX providers

---

## Versioning

This project follows semantic versioning.

Current stage: `v0.x` (early development)

---

## License

MIT License
