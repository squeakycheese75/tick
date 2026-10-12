package domain

import (
	"fmt"
	"strings"
	"time"
)

// TradeSide is the direction of a trade.
type TradeSide string

// Trade sides.
const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

// quantityEpsilon absorbs floating point noise when a sell closes a position.
const quantityEpsilon = 1e-9

// TradeInput is a single executed trade to apply to a portfolio.
type TradeInput struct {
	Ref            string
	Date           time.Time
	Symbol         string
	Side           TradeSide
	Quantity       float64
	Price          float64
	Currency       string
	Fee            float64
	FXRate         float64
	Exchange       string
	InstrumentType string
	QuoteCurrency  string
}

// ImportTradesInput is a batch of trades for one portfolio.
type ImportTradesInput struct {
	PortfolioName string
	Trades        []TradeInput
}

// ImportTradesOutput summarises an import run.
type ImportTradesOutput struct {
	PortfolioName string
	Applied       int
	Skipped       int
}

// Normalize trims and canonicalises the textual fields.
func (t *TradeInput) Normalize() {
	t.Ref = strings.TrimSpace(t.Ref)
	t.Symbol = strings.ToUpper(strings.TrimSpace(t.Symbol))
	t.Side = TradeSide(strings.ToLower(strings.TrimSpace(string(t.Side))))
	t.Currency = strings.ToUpper(strings.TrimSpace(t.Currency))
	t.Exchange = strings.ToUpper(strings.TrimSpace(t.Exchange))
	t.InstrumentType = strings.ToLower(strings.TrimSpace(t.InstrumentType))
	t.QuoteCurrency = strings.ToUpper(strings.TrimSpace(t.QuoteCurrency))
}

// Validate checks a normalised trade.
func (t TradeInput) Validate() error {
	switch {
	case t.Symbol == "":
		return fmt.Errorf("symbol is required")
	case t.Side != TradeSideBuy && t.Side != TradeSideSell:
		return fmt.Errorf("side must be buy or sell, got %q", t.Side)
	case t.Date.IsZero():
		return fmt.Errorf("date is required")
	case t.Quantity <= 0:
		return fmt.Errorf("quantity must be greater than 0")
	case t.Price < 0:
		return fmt.Errorf("price must be 0 or greater")
	case t.Fee < 0:
		return fmt.Errorf("fee must be 0 or greater")
	case t.FXRate < 0:
		return fmt.Errorf("fx rate must be 0 or greater")
	case t.Currency == "":
		return fmt.Errorf("currency is required")
	}

	return nil
}

// Validate checks the whole batch, reporting the offending row.
func (in ImportTradesInput) Validate() error {
	if strings.TrimSpace(in.PortfolioName) == "" {
		return fmt.Errorf("portfolioName is required")
	}
	if len(in.Trades) == 0 {
		return fmt.Errorf("at least one trade is required")
	}

	for i, t := range in.Trades {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("trades[%d]: %w", i, err)
		}
	}

	return nil
}

// TradePosition is a quantity held at an average cost per unit. AvgCost is in the instrument's
// quote currency; AvgCostBase is in the portfolio's base currency and is 0 when unknown.
type TradePosition struct {
	Quantity    float64
	AvgCost     float64
	AvgCostBase float64
}

// ApplyBuy adds quantity for a total cost, fees included, in the quote currency and in the
// base currency. A zero costBase, or an existing holding with no known base cost, leaves the
// base cost unknown rather than mixing known and unknown amounts.
func (h TradePosition) ApplyBuy(quantity, cost, costBase float64) TradePosition {
	total := h.Quantity + quantity

	avgBase := 0.0
	if costBase > 0 && (h.Quantity == 0 || h.AvgCostBase > 0) {
		avgBase = (h.Quantity*h.AvgCostBase + costBase) / total
	}

	return TradePosition{
		Quantity:    total,
		AvgCost:     (h.Quantity*h.AvgCost + cost) / total,
		AvgCostBase: avgBase,
	}
}

// ApplySell removes quantity at the existing average costs.
// Selling more than is held returns ErrInsufficientQuantity.
func (h TradePosition) ApplySell(quantity float64) (TradePosition, error) {
	remaining := h.Quantity - quantity
	if remaining < -quantityEpsilon {
		return TradePosition{}, fmt.Errorf(
			"sell %g but only %g held: %w",
			quantity,
			h.Quantity,
			ErrInsufficientQuantity,
		)
	}
	if remaining < quantityEpsilon {
		return TradePosition{}, nil
	}

	return TradePosition{Quantity: remaining, AvgCost: h.AvgCost, AvgCostBase: h.AvgCostBase}, nil
}

// IsClosed reports whether nothing is held.
func (h TradePosition) IsClosed() bool {
	return h.Quantity <= 0
}
