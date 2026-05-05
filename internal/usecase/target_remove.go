package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type RemoveTargetUsecase struct {
	targets    TargetRepository
	portfolios PortfolioRepository
}

func NewRemoveTargetUsecase(portfolioRepo PortfolioRepository, targetRepo TargetRepository) *RemoveTargetUsecase {
	return &RemoveTargetUsecase{
		portfolios: portfolioRepo,
		targets:    targetRepo,
	}
}

func (uc *RemoveTargetUsecase) Execute(ctx context.Context, in domain.DeleteTargetUseCaseInput) error {
	portfolio, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return fmt.Errorf("portfolio %q not found", in.PortfolioName)
		}
		return fmt.Errorf("get portfolio %q: %w", in.PortfolioName, err)
	}

	if in.TargetID <= 0 {
		return domain.ErrTargetNotFound
	}

	return uc.targets.Delete(ctx, in.TargetID, portfolio.ID)
}
