package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type portfolioLister interface {
	List(ctx context.Context) ([]repository.PortfolioOverview, error)
}

// ListPortfoliosUseCase lists every portfolio with its position count.
type ListPortfoliosUseCase struct {
	portfolios portfolioLister
}

// NewListPortfoliosUseCase builds a ListPortfoliosUseCase.
func NewListPortfoliosUseCase(portfolios portfolioLister) *ListPortfoliosUseCase {
	return &ListPortfoliosUseCase{portfolios: portfolios}
}

// Execute returns the portfolios in the order the store gives them.
func (uc *ListPortfoliosUseCase) Execute(ctx context.Context) (domain.ListPortfoliosOutput, error) {
	rows, err := uc.portfolios.List(ctx)
	if err != nil {
		return domain.ListPortfoliosOutput{}, fmt.Errorf("list portfolios: %w", err)
	}

	items := make([]domain.PortfolioListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, domain.PortfolioListItem{
			Name:          r.Name,
			BaseCurrency:  r.BaseCurrency,
			PositionCount: r.PositionCount,
		})
	}

	return domain.ListPortfoliosOutput{Portfolios: items}, nil
}
