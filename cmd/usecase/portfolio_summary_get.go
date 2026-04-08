package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/squeakycheese75/tick/internal/domain"
)

type (
	PortfolioRepository interface {
		GetByName(ctx context.Context, name string) (domain.Portfolio, error)
		Create(ctx context.Context, p domain.Portfolio) error
	}
	PositionRepository interface {
		ListByPortfolio(ctx context.Context, portfolioName string) ([]domain.Position, error)
		Create(ctx context.Context, p domain.Position) error
	}
	PriceProvider interface {
		GetPrice(ctx context.Context, ticker string) (price float64, currency string, err error)
	}
	FXProvider interface {
		GetRate(ctx context.Context, from string, to string) (float64, error)
	}
)

type GetPortfolioSummaryUseCase struct {
	portfolios PortfolioRepository
	positions  PositionRepository
	prices     PriceProvider
	fx         FXProvider
}

func NewGetPortfolioSummaryUseCase(portfolioRepo PortfolioRepository, positionRepo PositionRepository, prices PriceProvider, fx FXProvider) *GetPortfolioSummaryUseCase {
	return &GetPortfolioSummaryUseCase{
		portfolios: portfolioRepo,
		positions:  positionRepo,
		prices:     prices,
		fx:         fx,
	}
}

func (uc *GetPortfolioSummaryUseCase) Execute(ctx context.Context, in GetPortfolioSummaryUsecaseInput) (GetPortfolioSummaryUsecaseOutput, error) {
	pf, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		pf = domain.Portfolio{
			Name:         in.PortfolioName,
			BaseCurrency: basePortfolioCcy,
		}

		if err := uc.portfolios.Create(ctx, pf); err != nil {
			return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("create default portfolio: %w", err)
		}
	}

	positions, err := uc.positions.ListByPortfolio(ctx, in.PortfolioName)
	if err != nil {
		return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("list positions: %w", err)
	}

	result := GetPortfolioSummaryUsecaseOutput{
		PortfolioName: pf.Name,
		BaseCurrency:  pf.BaseCurrency,
		Positions:     make([]SummaryPosition, 0, len(positions)),
	}

	for _, pos := range positions {
		currentPrice, priceCurrency, err := uc.prices.GetPrice(ctx, pos.Ticker)
		if err != nil {
			return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("get price for %s: %w", pos.Ticker, err)
		}

		if priceCurrency == "" {
			priceCurrency = pos.InstrumentCurrency
		}

		fxRate, err := uc.fx.GetRate(ctx, priceCurrency, pf.BaseCurrency)
		if err != nil {
			return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("get fx rate %s/%s: %w", priceCurrency, pf.BaseCurrency, err)
		}

		marketValueBase := pos.Quantity * currentPrice * fxRate
		result.TotalValue += marketValueBase

		result.Positions = append(result.Positions, SummaryPosition{
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
