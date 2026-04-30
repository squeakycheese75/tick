package usecase

import (
	"context"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type SetTargetUseCase struct {
	portfolios PortfolioRepository
	targets    TargetRepository
}

func NewSetTargetUseCase(portfolios PortfolioRepository, targets TargetRepository) *SetTargetUseCase {
	return &SetTargetUseCase{
		portfolios: portfolios,
		targets:    targets,
	}
}

func (uc *SetTargetUseCase) Execute(ctx context.Context, in domain.SetTargetUseCaseInput) (*domain.SetTargetUseCaseOutput, error) {
	portfolio, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		return nil, err
	}

	target := domain.Target{
		PortfolioID:   portfolio.ID,
		Symbol:        strings.ToUpper(in.Symbol),
		Type:          in.Type,
		TargetPrice:   in.TargetPrice,
		QuoteCurrency: strings.ToUpper(in.QuoteCurrency),
	}

	return &domain.SetTargetUseCaseOutput{
		PortfolioName: in.PortfolioName,
		Symbol:        in.Symbol,
		QuoteCurrency: in.QuoteCurrency,
		Type:          in.Type,
		TargetPrice:   in.TargetPrice,
	}, uc.targets.Save(ctx, target)
}
