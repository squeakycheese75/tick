package domain_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

func TestTradePosition_ApplyBuy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		start    domain.TradePosition
		qty      float64
		cost     float64
		costBase float64
		want     domain.TradePosition
	}{
		{"opens a position", domain.TradePosition{}, 10, 1000, 900, domain.TradePosition{Quantity: 10, AvgCost: 100, AvgCostBase: 90}},
		{"averages local and base costs", domain.TradePosition{Quantity: 10, AvgCost: 100, AvgCostBase: 90}, 10, 2000, 1900, domain.TradePosition{Quantity: 20, AvgCost: 150, AvgCostBase: 140}},
		{"fees are part of the total cost", domain.TradePosition{}, 10, 1010, 909, domain.TradePosition{Quantity: 10, AvgCost: 101, AvgCostBase: 90.9}},
		{"unknown base cost on the new buy stays unknown", domain.TradePosition{Quantity: 10, AvgCost: 100, AvgCostBase: 90}, 10, 2000, 0, domain.TradePosition{Quantity: 20, AvgCost: 150}},
		{"unknown base cost on the existing holding stays unknown", domain.TradePosition{Quantity: 10, AvgCost: 100}, 10, 2000, 1900, domain.TradePosition{Quantity: 20, AvgCost: 150}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.start.ApplyBuy(tt.qty, tt.cost, tt.costBase)

			if got.Quantity != tt.want.Quantity || got.AvgCost != tt.want.AvgCost || math.Abs(got.AvgCostBase-tt.want.AvgCostBase) > 1e-9 {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestTradePosition_ApplySell(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		start      domain.TradePosition
		qty        float64
		want       domain.TradePosition
		wantClosed bool
		wantErr    error
	}{
		{"partial sell keeps average costs", domain.TradePosition{Quantity: 10, AvgCost: 100, AvgCostBase: 90}, 4, domain.TradePosition{Quantity: 6, AvgCost: 100, AvgCostBase: 90}, false, nil},
		{"selling everything closes", domain.TradePosition{Quantity: 10, AvgCost: 100}, 10, domain.TradePosition{}, true, nil},
		{"float noise still closes", domain.TradePosition{Quantity: 0.3, AvgCost: 1}, 0.1 + 0.2, domain.TradePosition{}, true, nil},
		{"overselling fails", domain.TradePosition{Quantity: 1, AvgCost: 100}, 2, domain.TradePosition{}, false, domain.ErrInsufficientQuantity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.start.ApplySell(tt.qty)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			if err == nil && got.IsClosed() != tt.wantClosed {
				t.Fatalf("IsClosed = %v, want %v", got.IsClosed(), tt.wantClosed)
			}
		})
	}
}

func TestImportTradesInput_Validate(t *testing.T) {
	t.Parallel()

	valid := domain.TradeInput{
		Date: time.Now(), Symbol: "AAPL", Side: domain.TradeSideBuy,
		Quantity: 1, Price: 10, Currency: "USD",
	}
	with := func(f func(*domain.TradeInput)) domain.TradeInput {
		tr := valid
		f(&tr)
		return tr
	}

	tests := []struct {
		name    string
		in      domain.ImportTradesInput
		wantErr bool
	}{
		{"valid", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{valid}}, false},
		{"missing portfolio", domain.ImportTradesInput{Trades: []domain.TradeInput{valid}}, true},
		{"no trades", domain.ImportTradesInput{PortfolioName: "p"}, true},
		{"missing symbol", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.Symbol = "" })}}, true},
		{"bad side", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.Side = "hold" })}}, true},
		{"missing date", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.Date = time.Time{} })}}, true},
		{"zero quantity", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.Quantity = 0 })}}, true},
		{"negative price", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.Price = -1 })}}, true},
		{"negative fee", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.Fee = -1 })}}, true},
		{"negative fx", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.FXRate = -1 })}}, true},
		{"missing currency", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{with(func(t *domain.TradeInput) { t.Currency = "" })}}, true},
		{"the same ref twice in one file", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{
			with(func(t *domain.TradeInput) { t.Ref = "order-1" }),
			with(func(t *domain.TradeInput) { t.Ref = "order-1" }),
		}}, true},
		{"different refs", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{
			with(func(t *domain.TradeInput) { t.Ref = "order-1" }),
			with(func(t *domain.TradeInput) { t.Ref = "order-2" }),
		}}, false},
		{"blank refs are not duplicates of each other", domain.ImportTradesInput{PortfolioName: "p", Trades: []domain.TradeInput{valid, valid}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.in.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTradeInput_Normalize(t *testing.T) {
	t.Parallel()

	tr := domain.TradeInput{
		Ref: " r1 ", Symbol: " aapl ", Side: " BUY ", Currency: " eur ",
		Exchange: " nasdaq ", InstrumentType: " EQUITY ", QuoteCurrency: " usd ",
	}
	tr.Normalize()

	want := domain.TradeInput{
		Ref: "r1", Symbol: "AAPL", Side: domain.TradeSideBuy, Currency: "EUR",
		Exchange: "NASDAQ", InstrumentType: "equity", QuoteCurrency: "USD",
	}
	if tr != want {
		t.Fatalf("got %+v, want %+v", tr, want)
	}
}

func TestPortfolioAnalysis_PricesAsOf(t *testing.T) {
	t.Parallel()

	fri := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	mon := time.Date(2026, 10, 12, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		positions []domain.AnalyzedPosition
		want      time.Time
	}{
		{"no positions", nil, time.Time{}},
		{"no provider timestamps", []domain.AnalyzedPosition{{}, {}}, time.Time{}},
		{"oldest of several", []domain.AnalyzedPosition{{QuoteAsOf: mon}, {QuoteAsOf: fri}, {}}, fri},
		{"order does not matter", []domain.AnalyzedPosition{{QuoteAsOf: fri}, {QuoteAsOf: mon}}, fri},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := domain.PortfolioAnalysis{AnalyzedPositions: tt.positions}.PricesAsOf()
			if !got.Equal(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
