package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
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
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioAlreadyExists) {
			return nil, fmt.Errorf("portfolio %q already exists", in.PortfolioName)
		}
		return nil, err
	}

	if !errors.Is(err, domain.ErrPortfolioNotFound) {
		return nil, fmt.Errorf("get portfolio: %w", err)
	}

	if err := uc.portfolios.Create(ctx, repository.Portfolio{
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
