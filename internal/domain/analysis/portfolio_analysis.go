package analysis

import (
	"context"
	"fmt"
	"sort"

	"github.com/squeakycheese75/tick/internal/domain"
)

type (
	PriceProvider interface {
		GetPrice(ctx context.Context, ticker string) (price float64, currency string, err error)
	}
	FXProvider interface {
		GetRate(ctx context.Context, from string, to string) (float64, error)
	}
)

type PortfolioAnalyzer struct {
	prices PriceProvider
	fx     FXProvider
}

func NewPortfolioAnalyzer(prices PriceProvider, fx FXProvider) *PortfolioAnalyzer {
	return &PortfolioAnalyzer{
		prices: prices,
		fx:     fx,
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
	CurrentPrice       float64
	FXRate             float64
	MarketValueBase    float64
	Weight             float64
	PriceCurrency      string
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
		currentPrice, priceCurrency, err := a.prices.GetPrice(ctx, pos.Ticker)
		if err != nil {
			return PortfolioAnalysis{}, fmt.Errorf("get price for %s: %w", pos.Ticker, err)
		}

		if priceCurrency == "" {
			priceCurrency = pos.InstrumentCurrency
		}

		fxRate, err := a.fx.GetRate(ctx, priceCurrency, in.Portfolio.BaseCurrency)
		if err != nil {
			return PortfolioAnalysis{}, fmt.Errorf("get fx rate %s/%s: %w", priceCurrency, in.Portfolio.BaseCurrency, err)
		}

		marketValueBase := pos.Quantity * currentPrice * fxRate
		result.TotalValue += marketValueBase

		result.AnalyzedPositions = append(result.AnalyzedPositions, AnalyzedPosition{
			Ticker:             pos.Ticker,
			Quantity:           pos.Quantity,
			AvgCost:            pos.AvgCost,
			InstrumentCurrency: pos.InstrumentCurrency,
			PriceCurrency:      priceCurrency,
			CurrentPrice:       currentPrice,
			FXRate:             fxRate,
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
