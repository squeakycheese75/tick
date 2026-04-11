package analysis

import (
	"context"
	"fmt"
	"sort"

	"github.com/squeakycheese75/tick/internal/domain"
)

type (
	PricingSvc interface {
		GetValuationQuote(ctx context.Context, ticker string, targetCurrency string, instrumentCurrency string) (domain.ValuationQuote, error)
	}
)

type PortfolioAnalyzer struct {
	pricingSvc PricingSvc
}

func NewPortfolioAnalyzer(pricingSvc PricingSvc) *PortfolioAnalyzer {
	return &PortfolioAnalyzer{
		pricingSvc: pricingSvc,
	}
}

type AnalyzePortfolioInput struct {
	Portfolio domain.Portfolio
	Positions []domain.Position
}

type AnalyzedPosition struct {
	Symbol             string
	Quantity           float64
	AvgCost            float64
	InstrumentCurrency string
	QuotedPrice        float64
	QuotedChange       float64
	QuotedChangePct    float64
	PriceCurrency      string
	FXRate             float64
	ConvertedPrice     float64
	ConvertedChange    float64
	MarketValueBase    float64
	Weight             float64
}

type PortfolioAnalysis struct {
	PortfolioName     string
	BaseCurrency      string
	AnalyzedPositions []AnalyzedPosition
	TotalValue        float64
}

func (a *PortfolioAnalyzer) Analyze(ctx context.Context, in AnalyzePortfolioInput) (PortfolioAnalysis, error) {
	result := PortfolioAnalysis{
		PortfolioName:     in.Portfolio.Name,
		BaseCurrency:      in.Portfolio.BaseCurrency,
		AnalyzedPositions: make([]AnalyzedPosition, 0, len(in.Positions)),
	}

	for _, pos := range in.Positions {
		valuationQuote, err := a.pricingSvc.GetValuationQuote(ctx, pos.Instrument.Symbol, in.Portfolio.BaseCurrency, pos.Instrument.QuoteCurrency)
		if err != nil {
			return PortfolioAnalysis{}, fmt.Errorf("get valuation quote for %s: %w", pos.Instrument.Symbol, err)
		}

		marketValueBase := pos.Quantity * valuationQuote.ConvertedPrice
		result.TotalValue += marketValueBase

		result.AnalyzedPositions = append(result.AnalyzedPositions, AnalyzedPosition{
			Symbol:             pos.Instrument.Symbol,
			Quantity:           pos.Quantity,
			AvgCost:            pos.AvgCost,
			InstrumentCurrency: pos.Instrument.QuoteCurrency,
			QuotedPrice:        valuationQuote.Quote.Price,
			QuotedChange:       valuationQuote.Quote.Change,
			QuotedChangePct:    valuationQuote.Quote.ChangePercent,
			PriceCurrency:      valuationQuote.Quote.PriceCurrency,
			FXRate:             valuationQuote.FXRate,
			ConvertedPrice:     valuationQuote.ConvertedPrice,
			MarketValueBase:    marketValueBase,
		})
	}

	if result.TotalValue > 0 {
		for i := range result.AnalyzedPositions {
			result.AnalyzedPositions[i].Weight = result.AnalyzedPositions[i].MarketValueBase / result.TotalValue
		}
	}

	sort.Slice(result.AnalyzedPositions, func(i, j int) bool {
		return result.AnalyzedPositions[i].MarketValueBase > result.AnalyzedPositions[j].MarketValueBase
	})

	return result, nil
}
