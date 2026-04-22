package analysis

import (
	"context"
	"errors"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis/mocks"
	"go.uber.org/mock/gomock"
)

func TestPortfolioAnalyzer_Analyze_SinglePosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pricingSvc := mocks.NewMockPricingSvc(ctrl)
	analyzer := NewPortfolioAnalyzer(pricingSvc)

	in := AnalyzePortfolioInput{
		Portfolio: domain.Portfolio{
			Name:         "main",
			BaseCurrency: "EUR",
		},
		Positions: []domain.Position{
			{
				Quantity: 10,
				AvgCost:  400,
				Instrument: domain.Instrument{
					Symbol:        "NVDA",
					QuoteCurrency: "USD",
					AssetType:     "equity",
				},
			},
		},
	}

	pricingSvc.EXPECT().
		GetValuationQuote(gomock.Any(), "NVDA", "EUR", "USD", "equity").
		Return(domain.ValuationQuote{
			Quote: domain.Quote{
				Price:         450,
				Change:        5,
				ChangePercent: 1.12,
				PriceCurrency: "USD",
			},
			FXRate:         0.92,
			ConvertedPrice: 414,
		}, nil)

	out, err := analyzer.Analyze(context.Background(), in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out.PortfolioName != "main" {
		t.Fatalf("unexpected portfolio name: %q", out.PortfolioName)
	}

	if out.BaseCurrency != "EUR" {
		t.Fatalf("unexpected base currency: %q", out.BaseCurrency)
	}

	if len(out.AnalyzedPositions) != 1 {
		t.Fatalf("expected 1 analyzed position, got %d", len(out.AnalyzedPositions))
	}

	got := out.AnalyzedPositions[0]

	if got.Symbol != "NVDA" {
		t.Fatalf("unexpected symbol: %q", got.Symbol)
	}

	if got.Quantity != 10 {
		t.Fatalf("unexpected quantity: %v", got.Quantity)
	}

	if got.AvgCost != 400 {
		t.Fatalf("unexpected avg cost: %v", got.AvgCost)
	}

	if got.InstrumentCurrency != "USD" {
		t.Fatalf("unexpected instrument currency: %q", got.InstrumentCurrency)
	}

	if got.QuotedPrice != 450 {
		t.Fatalf("unexpected quoted price: %v", got.QuotedPrice)
	}

	if got.QuotedChange != 5 {
		t.Fatalf("unexpected quoted change: %v", got.QuotedChange)
	}

	if got.QuotedChangePct != 1.12 {
		t.Fatalf("unexpected quoted change pct: %v", got.QuotedChangePct)
	}

	if got.PriceCurrency != "USD" {
		t.Fatalf("unexpected price currency: %q", got.PriceCurrency)
	}

	if got.FXRate != 0.92 {
		t.Fatalf("unexpected fx rate: %v", got.FXRate)
	}

	if got.ConvertedPrice != 414 {
		t.Fatalf("unexpected converted price: %v", got.ConvertedPrice)
	}

	wantMarketValue := 10 * 414.0
	if got.MarketValueBase != wantMarketValue {
		t.Fatalf("unexpected market value: got %v want %v", got.MarketValueBase, wantMarketValue)
	}

	if out.TotalValue != wantMarketValue {
		t.Fatalf("unexpected total value: got %v want %v", out.TotalValue, wantMarketValue)
	}

	if got.Weight != 1 {
		t.Fatalf("unexpected weight: got %v want 1", got.Weight)
	}
}

func TestPortfolioAnalyzer_Analyze_MultiplePositions_SortsAndCalculatesWeights(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pricingSvc := mocks.NewMockPricingSvc(ctrl)
	analyzer := NewPortfolioAnalyzer(pricingSvc)

	in := AnalyzePortfolioInput{
		Portfolio: domain.Portfolio{
			Name:         "main",
			BaseCurrency: "EUR",
		},
		Positions: []domain.Position{
			{
				Quantity: 2,
				AvgCost:  100,
				Instrument: domain.Instrument{
					Symbol:        "AAA",
					QuoteCurrency: "USD",
					AssetType:     "equity",
				},
			},
			{
				Quantity: 5,
				AvgCost:  50,
				Instrument: domain.Instrument{
					Symbol:        "BBB",
					QuoteCurrency: "USD",
					AssetType:     "equity",
				},
			},
		},
	}

	pricingSvc.EXPECT().
		GetValuationQuote(gomock.Any(), "AAA", "EUR", "USD", "equity").
		Return(domain.ValuationQuote{
			Quote: domain.Quote{
				Price:         100,
				Change:        2,
				ChangePercent: 2,
				PriceCurrency: "USD",
			},
			FXRate:         1,
			ConvertedPrice: 100,
		}, nil)

	pricingSvc.EXPECT().
		GetValuationQuote(gomock.Any(), "BBB", "EUR", "USD", "equity").
		Return(domain.ValuationQuote{
			Quote: domain.Quote{
				Price:         50,
				Change:        1,
				ChangePercent: 2,
				PriceCurrency: "USD",
			},
			FXRate:         1,
			ConvertedPrice: 50,
		}, nil)

	out, err := analyzer.Analyze(context.Background(), in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(out.AnalyzedPositions) != 2 {
		t.Fatalf("expected 2 positions, got %d", len(out.AnalyzedPositions))
	}

	// AAA = 2 * 100 = 200
	// BBB = 5 * 50  = 250
	if out.TotalValue != 450 {
		t.Fatalf("unexpected total value: got %v want 450", out.TotalValue)
	}

	// Sorted descending by market value, so BBB first.
	if out.AnalyzedPositions[0].Symbol != "BBB" {
		t.Fatalf("expected BBB first, got %q", out.AnalyzedPositions[0].Symbol)
	}

	if out.AnalyzedPositions[1].Symbol != "AAA" {
		t.Fatalf("expected AAA second, got %q", out.AnalyzedPositions[1].Symbol)
	}

	if out.AnalyzedPositions[0].MarketValueBase != 250 {
		t.Fatalf("unexpected BBB market value: %v", out.AnalyzedPositions[0].MarketValueBase)
	}

	if out.AnalyzedPositions[1].MarketValueBase != 200 {
		t.Fatalf("unexpected AAA market value: %v", out.AnalyzedPositions[1].MarketValueBase)
	}

	if out.AnalyzedPositions[0].Weight != 250.0/450.0 {
		t.Fatalf("unexpected BBB weight: got %v want %v", out.AnalyzedPositions[0].Weight, 250.0/450.0)
	}

	if out.AnalyzedPositions[1].Weight != 200.0/450.0 {
		t.Fatalf("unexpected AAA weight: got %v want %v", out.AnalyzedPositions[1].Weight, 200.0/450.0)
	}
}

func TestPortfolioAnalyzer_Analyze_ReturnsErrorWhenPricingFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pricingSvc := mocks.NewMockPricingSvc(ctrl)
	analyzer := NewPortfolioAnalyzer(pricingSvc)

	in := AnalyzePortfolioInput{
		Portfolio: domain.Portfolio{
			Name:         "main",
			BaseCurrency: "EUR",
		},
		Positions: []domain.Position{
			{
				Quantity: 1,
				Instrument: domain.Instrument{
					Symbol:        "NVDA",
					QuoteCurrency: "USD",
					AssetType:     "equity",
				},
			},
		},
	}

	pricingSvc.EXPECT().
		GetValuationQuote(gomock.Any(), "NVDA", "EUR", "USD", "equity").
		Return(domain.ValuationQuote{}, errors.New("pricing failure"))

	_, err := analyzer.Analyze(context.Background(), in)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPortfolioAnalyzer_Analyze_ZeroTotalValue_LeavesWeightsAtZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pricingSvc := mocks.NewMockPricingSvc(ctrl)
	analyzer := NewPortfolioAnalyzer(pricingSvc)

	in := AnalyzePortfolioInput{
		Portfolio: domain.Portfolio{
			Name:         "main",
			BaseCurrency: "EUR",
		},
		Positions: []domain.Position{
			{
				Quantity: 10,
				Instrument: domain.Instrument{
					Symbol:        "ZERO",
					QuoteCurrency: "USD",
					AssetType:     "equity",
				},
			},
		},
	}

	pricingSvc.EXPECT().
		GetValuationQuote(gomock.Any(), "ZERO", "EUR", "USD", "equity").
		Return(domain.ValuationQuote{
			Quote: domain.Quote{
				Price:         0,
				Change:        0,
				ChangePercent: 0,
				PriceCurrency: "USD",
			},
			FXRate:         1,
			ConvertedPrice: 0,
		}, nil)

	out, err := analyzer.Analyze(context.Background(), in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out.TotalValue != 0 {
		t.Fatalf("unexpected total value: %v", out.TotalValue)
	}

	if len(out.AnalyzedPositions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(out.AnalyzedPositions))
	}

	if out.AnalyzedPositions[0].Weight != 0 {
		t.Fatalf("expected zero weight, got %v", out.AnalyzedPositions[0].Weight)
	}
}

func TestPortfolioAnalyzer_Analyze_SetsConvertedChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pricingSvc := mocks.NewMockPricingSvc(ctrl)
	analyzer := NewPortfolioAnalyzer(pricingSvc)

	in := AnalyzePortfolioInput{
		Portfolio: domain.Portfolio{
			Name:         "main",
			BaseCurrency: "EUR",
		},
		Positions: []domain.Position{
			{
				Quantity: 10,
				AvgCost:  100,
				Instrument: domain.Instrument{
					Symbol:        "NVDA",
					QuoteCurrency: "USD",
					AssetType:     "equity",
				},
			},
		},
	}

	pricingSvc.EXPECT().
		GetValuationQuote(gomock.Any(), "NVDA", "EUR", "USD", "equity").
		Return(domain.ValuationQuote{
			Quote: domain.Quote{
				Price:         200,
				Change:        5,
				ChangePercent: 2.5,
				PriceCurrency: "USD",
			},
			FXRate:         0.9,
			ConvertedPrice: 180,
		}, nil)

	out, err := analyzer.Analyze(context.Background(), in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(out.AnalyzedPositions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(out.AnalyzedPositions))
	}

	got := out.AnalyzedPositions[0]

	want := 5 * 0.9

	if got.ConvertedChange != want {
		t.Fatalf("unexpected converted change: got %v want %v", got.ConvertedChange, want)
	}
}
