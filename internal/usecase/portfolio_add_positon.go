package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/repository"
)

type AddPositionToPortfolioUseCase struct {
	portfolios  PortfolioRepository
	positions   PositionRepository
	instruments InstrumentRepository
}

func NewAddPositionToPortfolioUseCase(positionRepo PositionRepository, portfolioRepo PortfolioRepository, instrumentRepo InstrumentRepository) *AddPositionToPortfolioUseCase {
	return &AddPositionToPortfolioUseCase{
		positions:   positionRepo,
		portfolios:  portfolioRepo,
		instruments: instrumentRepo,
	}
}

func (uc *AddPositionToPortfolioUseCase) Execute(
	ctx context.Context,
	in AddPositionToPortfolioInput,
) (*AddPositionToPortfolioOutput, error) {
	portfolio, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		return nil, fmt.Errorf("get portfolio %q: %w", in.PortfolioName, err)
	}

	instrument, err := uc.instruments.GetBySymbol(ctx, in.Symbol)
	if err != nil {
		return nil, fmt.Errorf("get instrument %q: %w", in.Symbol, err)
	}

	err = uc.positions.Create(
		ctx,
		repository.CreatePositionParams{
			InstrumentID: instrument.ID,
			PortfolioID:  portfolio.ID,
			Quantity:     in.Qty,
			AvgCost:      in.AvgCost,
			Currency:     in.QuoteCurrency,
		},
	)
	if err != nil {
		return nil, err
	}

	return &AddPositionToPortfolioOutput{}, err
}
