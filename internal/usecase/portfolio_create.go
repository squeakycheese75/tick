package usecase

import (
	"context"
	"errors"
	"fmt"

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
	_, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err == nil {
		// portfolio already exists
		return nil, fmt.Errorf(
			"portfolio %q already exists. Use a different name or update it",
			in.PortfolioName,
		)
	}
	if !errors.Is(err, domain.ErrPortfolioNotFound) {
		// real error
		return nil, fmt.Errorf("get portfolio: %w", err)
	}

	if err := uc.portfolios.Create(ctx, domain.Portfolio{
		Name:         in.PortfolioName,
		BaseCurrency: in.BaseCurrency,
	}); err != nil {
		return nil, err
	}

	return &CreatePortfolioUsecaseOutout{
		PortfolioName: in.PortfolioName,
		BaseCurrency:  in.BaseCurrency,
	}, nil
}
