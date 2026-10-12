package usecase_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/usecase"
)

type fakeAnalysis struct {
	result domain.PortfolioAnalysis
	err    error
}

func (f fakeAnalysis) GetAnalysis(context.Context, string) (domain.PortfolioAnalysis, error) {
	return f.result, f.err
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestGetPortfolioSummary_Execute_SplitsPnLIntoAssetAndFX(t *testing.T) {
	t.Parallel()

	// Bought 10 units at 100 USD when 1 USD = 0.90 EUR (cost 900 EUR).
	// Now 1 USD = 0.80 EUR and the price is 110 USD (value 880 EUR).
	// Total -20 EUR = asset +80 (10 x 10 USD x 0.80) and FX -100 (1000 USD x -0.10).
	tracked := domain.AnalyzedPosition{
		Symbol: "AAPL", Quantity: 10, AvgCost: 100, AvgCostBase: 90, FXRate: 0.8,
		QuotedPrice: 110, MarketValueBase: 880, InstrumentCurrency: "USD",
	}
	// No recorded purchase rate: cost taken at today's rate, all PnL reported as asset PnL.
	untracked := domain.AnalyzedPosition{
		Symbol: "MSFT", Quantity: 1, AvgCost: 100, FXRate: 0.8,
		QuotedPrice: 110, MarketValueBase: 88, InstrumentCurrency: "USD",
	}

	tests := []struct {
		name          string
		positions     []domain.AnalyzedPosition
		wantCost      float64
		wantPnL       float64
		wantAsset     float64
		wantFX        float64
		wantFirstKnow bool
	}{
		{"tracked position splits", []domain.AnalyzedPosition{tracked}, 900, -20, 80, -100, true},
		{"untracked position is all asset and excluded from the FX total", []domain.AnalyzedPosition{untracked}, 80, 8, 0, 0, false},
		{"mixed positions only split the tracked one", []domain.AnalyzedPosition{tracked, untracked}, 980, -12, 80, -100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := usecase.NewGetPortfolioSummaryUseCase(fakeAnalysis{result: domain.PortfolioAnalysis{
				PortfolioName: "p", BaseCurrency: "EUR", AnalyzedPositions: tt.positions,
			}})

			out, err := uc.Execute(context.Background(), domain.GetPortfolioSummaryUsecaseInput{PortfolioName: "p"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			r := out.Report
			if !near(r.TotalCost, tt.wantCost) || !near(r.TotalPnL, tt.wantPnL) ||
				!near(r.TotalAssetPnL, tt.wantAsset) || !near(r.TotalFXPnL, tt.wantFX) {
				t.Fatalf("totals = cost %v pnl %v asset %v fx %v, want %v %v %v %v",
					r.TotalCost, r.TotalPnL, r.TotalAssetPnL, r.TotalFXPnL,
					tt.wantCost, tt.wantPnL, tt.wantAsset, tt.wantFX)
			}
			if r.Positions[0].FXKnown != tt.wantFirstKnow {
				t.Fatalf("first position FXKnown = %v, want %v", r.Positions[0].FXKnown, tt.wantFirstKnow)
			}
		})
	}
}

func TestGetPortfolioSummary_Execute_PositionPnLAddsUp(t *testing.T) {
	t.Parallel()

	uc := usecase.NewGetPortfolioSummaryUseCase(fakeAnalysis{result: domain.PortfolioAnalysis{
		PortfolioName: "p", BaseCurrency: "EUR",
		AnalyzedPositions: []domain.AnalyzedPosition{{
			Symbol: "AAPL", Quantity: 10, AvgCost: 100, AvgCostBase: 90, FXRate: 0.8, MarketValueBase: 880,
		}},
	}})

	out, err := uc.Execute(context.Background(), domain.GetPortfolioSummaryUsecaseInput{PortfolioName: "p"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := out.Report.Positions[0]
	if !near(p.AssetPnL+p.FXPnL, p.UnrealizedPnL) {
		t.Fatalf("asset %v + fx %v != pnl %v", p.AssetPnL, p.FXPnL, p.UnrealizedPnL)
	}
	if !near(p.UnrealizedPnLPct, -20.0/900.0) {
		t.Fatalf("pnl pct = %v", p.UnrealizedPnLPct)
	}
}

func TestGetPortfolioSummary_Execute_ZeroCostHasNoPercent(t *testing.T) {
	t.Parallel()

	uc := usecase.NewGetPortfolioSummaryUseCase(fakeAnalysis{result: domain.PortfolioAnalysis{
		PortfolioName: "p", BaseCurrency: "EUR",
		AnalyzedPositions: []domain.AnalyzedPosition{{Symbol: "FREE", Quantity: 1, FXRate: 1, MarketValueBase: 5}},
	}})

	out, err := uc.Execute(context.Background(), domain.GetPortfolioSummaryUsecaseInput{PortfolioName: "p"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Report.TotalPnLPct != 0 || out.Report.Positions[0].UnrealizedPnLPct != 0 {
		t.Fatalf("expected zero percentages, got %+v", out.Report)
	}
}

func TestGetPortfolioSummary_Execute_AnalysisError(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	uc := usecase.NewGetPortfolioSummaryUseCase(fakeAnalysis{err: boom})

	if _, err := uc.Execute(context.Background(), domain.GetPortfolioSummaryUsecaseInput{}); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
