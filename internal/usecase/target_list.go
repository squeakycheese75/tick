package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type ListTargetsUseCase struct {
	portfolios PortfolioRepository
	targets    TargetRepository
}

func NewListTargetsUseCase(
	portfolios PortfolioRepository,
	targets TargetRepository,
) *ListTargetsUseCase {
	return &ListTargetsUseCase{
		portfolios: portfolios,
		targets:    targets,
	}
}

func (uc *ListTargetsUseCase) Execute(
	ctx context.Context,
	in domain.ListTargetsUseCaseInput,
) (domain.ListTargetsUseCaseOutput, error) {

	portfolio, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		return domain.ListTargetsUseCaseOutput{}, fmt.Errorf("get portfolio: %w", err)
	}

	targets, err := uc.targets.ListByPortfolio(ctx, portfolio.ID)
	if err != nil {
		return domain.ListTargetsUseCaseOutput{}, fmt.Errorf("list targets: %w", err)
	}

	return domain.ListTargetsUseCaseOutput{
		PortfolioName: in.PortfolioName,
		Targets:       targets,
	}, nil
}
