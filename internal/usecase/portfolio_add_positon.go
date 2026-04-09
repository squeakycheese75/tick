package usecase

import (
	"context"
	"errors"
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

func (uc *AddPositionToPortfolioUseCase) Execute(ctx context.Context, in AddPositionToPortfolioUseCaseInput) (AddPositionToPortfolioUseCaseOutput, error) {
	_, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return AddPositionToPortfolioUseCaseOutput{}, fmt.Errorf(
				"portfolio %q not found. Create it with:\n  tick portfolio create %s --base-currency EUR",
				in.PortfolioName,
				in.PortfolioName,
			)
		}
		return AddPositionToPortfolioUseCaseOutput{}, fmt.Errorf("get portfolio: %w", err)
	}

	if err := uc.positions.Create(ctx, domain.Position{
		PortfolioName:      in.PortfolioName,
		Ticker:             in.Ticker,
		Quantity:           in.Qty,
		AvgCost:            in.AvgCost,
		InstrumentCurrency: in.Currency,
	}); err != nil {
		return AddPositionToPortfolioUseCaseOutput{}, err
	}

	return AddPositionToPortfolioUseCaseOutput(in), nil
}
