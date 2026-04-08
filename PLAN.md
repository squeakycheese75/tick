# assetcli

A terminal-native portfolio, market research, and synthetic strategy engine.

## Overview

`assetcli` is a Bloomberg-style CLI for:

- Market research (stocks, ETFs, crypto)
- Portfolio tracking (real holdings)
- Synthetic portfolios (simulated strategies)
- AI-powered explanations and summaries

The system combines deterministic financial calculations with AI-driven interpretation.

---

## Core Principles

- **Local-first**: Data stored locally (SQLite)
- **Deterministic analytics**: All calculations done via code/SQL
- **AI for interpretation only**: Summaries, explanations, comparisons
- **Evidence-backed outputs**: Always show supporting data

---

## Features

### 1. Asset Research

- Asset overview
- News and catalysts
- Performance analysis
- Fundamentals and valuation
- AI-generated summaries

### 2. Portfolio Tracking

- Track real holdings
- Performance and PnL
- Exposure (sector, region)
- Contribution analysis

### 3. Synthetic Portfolios

- Simulated portfolios
- Strategy testing
- Rebalancing
- Scenario comparison

### 4. AI Layer

- Natural language queries
- Portfolio explanations
- Asset comparisons
- Report generation

---

## CLI Commands

### Asset Research

```bash
assetcli info <ticker>
assetcli compare <ticker...>
assetcli news <ticker>
assetcli price <ticker>
assetcli perf <ticker...>
assetcli chart <ticker>
assetcli fundamentals <ticker>
assetcli valuation <ticker>
assetcli financials <ticker>
assetcli brief <ticker>
assetcli ask "<question>"
```

### Portfolio

```bash
assetcli portfolio create <name>
assetcli portfolio add <ticker> --qty <n> --price <p>
assetcli portfolio remove <ticker> --qty <n>
assetcli portfolio summary
assetcli portfolio exposure
assetcli portfolio performance
assetcli portfolio contributors
assetcli portfolio ask "<question>"
```

### Synthetic Portfolios

```bash
assetcli sim create <name>
assetcli sim add <ticker> --weight <w>
assetcli sim remove <ticker>
assetcli sim summary <name>
assetcli sim performance <name>
assetcli sim rebalance <name>
```

### Comparison

```bash
assetcli compare portfolio <name>
assetcli compare sim <name>
assetcli compare portfolio <name> sim <name>
```

### Screening (future)

```bash
assetcli screen --sector <sector>
assetcli screen --market-cap <range>
assetcli screen --pe-under <value>
```

---

## Example Usage

```bash
assetcli info NVDA
assetcli compare SAP ASML NVDA
assetcli news NVDA

assetcli portfolio create main
assetcli portfolio add NVDA --qty 10 --price 400
assetcli portfolio summary

assetcli sim create ai-basket
assetcli sim add NVDA --weight 0.4
assetcli sim add ASML --weight 0.3
assetcli sim add TSM --weight 0.3

assetcli compare portfolio main sim ai-basket
```

---

## Architecture

```
cmd/assetcli
internal/
  ai/
  analytics/
  portfolio/
  sim/
  pricing/
  ingest/
  store/
  output/
```

### Storage

- SQLite (default)
- Optional Postgres (future)

---

## Data Sources (MVP)

- CSV (holdings, trades)
- Static asset metadata
- Optional price API

Future:
- Market data APIs
- News APIs
- Fundamentals APIs

---

## MVP Scope

### Daily workflow first

The project should start from the core daily workflow:

1. Check portfolio positions
2. Look for relevant news
3. Investigate an interesting stock
4. Review portfolio risk

These daily actions should define the first version of the CLI.

### v1

Focus on the smallest useful terminal workflow:

- `assetcli portfolio summary` — check current positions and allocation
- `assetcli news <ticker>` — review recent relevant news
- `assetcli info <ticker>` — investigate a stock quickly
- `assetcli risk` / `assetcli portfolio risk` — review portfolio concentration and basic risk
- `assetcli compare <ticker...>` — optional early comparison helper

### v2

- Portfolio performance and contributors
- Synthetic portfolios
- Portfolio vs synthetic comparison
- AI explanations and briefings

### v3

- Strategy rules
- Screening
- Advanced analytics
- Rebalancing workflows

---

## Key Questions the CLI Answers

### Daily workflow

- What do I hold right now?
- What moved in my portfolio?
- Is there important news on a stock I hold or watch?
- What does this company do and why is it interesting?
- Where is my portfolio risk concentrated?

### Asset

- What does this company do?
- What changed recently?
- What are the key risks?
- Is it expensive?
- How does it compare to peers?

### Portfolio

- What are my current positions?
- What is my exposure by asset, sector, and region?
- What drove returns?
- Am I too concentrated?
- What is the main risk in the portfolio?

### Synthetic

- How would this strategy perform?
- Is it better than my portfolio?

---

## Tech Stack

- Go (CLI + core engine)
- Cobra (CLI framework)
- SQLite (storage)
- Optional Python (future ML helpers)

---

## Next Steps

1. Bootstrap CLI with Cobra
2. Implement storage layer (SQLite)
3. Implement portfolio model and position storage
4. Build `assetcli portfolio summary`
5. Build `assetcli info <ticker>`
6. Build `assetcli news <ticker>`
7. Build `assetcli portfolio risk`
8. Add `compare` and `brief` after the core daily flow works

---

## Vision

A fast, local, developer-first alternative to Bloomberg-style tools, focused on:

- clarity
- control
- composability
- intelligent analysis

