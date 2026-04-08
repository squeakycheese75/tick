package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/squeakycheese75/tick/internal/domain"
)

type PortfolioRepository interface {
	GetByName(ctx context.Context, name string) (domain.Portfolio, error)
}

type PositionRepository interface {
	ListByPortfolio(ctx context.Context, portfolioName string) ([]domain.Position, error)
}

type PriceProvider interface {
	GetPrice(ctx context.Context, ticker string) (price float64, currency string, err error)
}

type FXProvider interface {
	GetRate(ctx context.Context, from string, to string) (float64, error)
}

type PortfolioService struct {
	portfolios PortfolioRepository
	positions  PositionRepository
	prices     PriceProvider
	fx         FXProvider
}

func NewPortfolioService(
	portfolios PortfolioRepository,
	positions PositionRepository,
	prices PriceProvider,
	fx FXProvider,
) *PortfolioService {
	return &PortfolioService{
		portfolios: portfolios,
		positions:  positions,
		prices:     prices,
		fx:         fx,
	}
}

func (s *PortfolioService) GetSummary(ctx context.Context, portfolioName string) (domain.Summary, error) {
	pf, err := s.portfolios.GetByName(ctx, portfolioName)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("get portfolio: %w", err)
	}

	positions, err := s.positions.ListByPortfolio(ctx, portfolioName)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("list positions: %w", err)
	}

	result := domain.Summary{
		PortfolioName: pf.Name,
		BaseCurrency:  pf.BaseCurrency,
		Positions:     make([]domain.SummaryPosition, 0, len(positions)),
	}

	for _, pos := range positions {
		currentPrice, priceCurrency, err := s.prices.GetPrice(ctx, pos.Ticker)
		if err != nil {
			return domain.Summary{}, fmt.Errorf("get price for %s: %w", pos.Ticker, err)
		}

		if priceCurrency == "" {
			priceCurrency = pos.InstrumentCurrency
		}

		fxRate, err := s.fx.GetRate(ctx, priceCurrency, pf.BaseCurrency)
		if err != nil {
			return domain.Summary{}, fmt.Errorf("get fx rate %s/%s: %w", priceCurrency, pf.BaseCurrency, err)
		}

		marketValueBase := pos.Quantity * currentPrice * fxRate
		result.TotalValue += marketValueBase

		result.Positions = append(result.Positions, domain.SummaryPosition{
			Ticker:             pos.Ticker,
			Quantity:           pos.Quantity,
			InstrumentCurrency: pos.InstrumentCurrency,
			BaseCurrency:       pf.BaseCurrency,
			AvgCost:            pos.AvgCost,
			CurrentPrice:       currentPrice,
			FXRate:             fxRate,
			MarketValueBase:    marketValueBase,
		})
	}

	if result.TotalValue > 0 {
		for i := range result.Positions {
			result.Positions[i].Weight = result.Positions[i].MarketValueBase / result.TotalValue
		}
	}

	sort.Slice(result.Positions, func(i, j int) bool {
		return result.Positions[i].MarketValueBase > result.Positions[j].MarketValueBase
	})

	return result, nil
}
