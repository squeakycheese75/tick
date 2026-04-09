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
	Ticker             string
	Quantity           float64
	AvgCost            float64
	InstrumentCurrency string
	QuotedPrice        float64
	QuoteCurrency      string
	FXRate             float64
	ConvertedPrice     float64
	BaseCurrency       string
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
		valuationQuote, err := a.pricingSvc.GetValuationQuote(ctx, pos.Ticker, in.Portfolio.BaseCurrency, pos.InstrumentCurrency)
		if err != nil {
			return PortfolioAnalysis{}, fmt.Errorf("get valuation quote for %s: %w", pos.Ticker, err)
		}

		marketValueBase := pos.Quantity * valuationQuote.ConvertedPrice
		result.TotalValue += marketValueBase

		result.AnalyzedPositions = append(result.AnalyzedPositions, AnalyzedPosition{
			Ticker:             pos.Ticker,
			Quantity:           pos.Quantity,
			AvgCost:            pos.AvgCost,
			InstrumentCurrency: pos.InstrumentCurrency,
			QuotedPrice:        valuationQuote.Quote.Price,
			QuoteCurrency:      valuationQuote.Quote.PriceCurrency,
			FXRate:             valuationQuote.FXRate,
			ConvertedPrice:     valuationQuote.ConvertedPrice,
			BaseCurrency:       valuationQuote.TargetCurrency,
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
