package usecase

import (
	"context"

	"github.com/squeakycheese75/tick/internal/domain"
)

type CreatePortfolioUseCase struct {
	portfolios PortfolioRepository
}

func NewCreatePortfolioUseCase(portfolioRepo PortfolioRepository) *CreatePortfolioUseCase {
	return &CreatePortfolioUseCase{
		portfolios: portfolioRepo,
	}
}

func (uc *CreatePortfolioUseCase) Execute(ctx context.Context, in CreatePortfolioUsecaseInput) (*CreatePortfolioUsecaseOutout, error) {
	if err := uc.portfolios.Create(ctx, domain.Portfolio{
		Name:         in.Name,
		BaseCurrency: in.BaseCurrency,
	}); err != nil {
		return nil, err
	}

	return &CreatePortfolioUsecaseOutout{
		Name:         in.Name,
		BaseCurrency: in.BaseCurrency,
	}, nil
}
