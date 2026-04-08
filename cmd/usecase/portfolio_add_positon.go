package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type AddPositionToPortfolioUseCase struct {
	portfolios PortfolioRepository
	positions  PositionRepository
}

func NewAddPositionToPortfolioUseCase(positionRepo PositionRepository, portfolioRepo PortfolioRepository) *AddPositionToPortfolioUseCase {
	return &AddPositionToPortfolioUseCase{
		positions:  positionRepo,
		portfolios: portfolioRepo,
	}
}

func (uc *AddPositionToPortfolioUseCase) Execute(ctx context.Context, in AddPositionToPortfolioUseCaseInput) (*AddPositionToPortfolioUseCaseOutput, error) {
	_, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		return nil, fmt.Errorf("portfolio: %w doesn't exist", err)
	}

	if err := uc.positions.Create(context.Background(), domain.Position{
		PortfolioName:      in.PortfolioName,
		Ticker:             in.Ticker,
		Quantity:           in.Qty,
		AvgCost:            in.AvgCost,
		InstrumentCurrency: in.Currency,
	}); err != nil {
		return nil, err
	}

	return &AddPositionToPortfolioUseCaseOutput{
		PortfolioName: in.PortfolioName,
		Ticker:        in.Ticker,
		Qty:           in.Qty,
		AvgCost:       in.AvgCost,
		Currency:      in.Currency,
	}, nil
}
