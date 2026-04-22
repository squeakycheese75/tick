package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type GetPortfolioRiskUseCase struct {
	portfolioSvc PortfolioSvc
}

func NewGetPortfolioRiskUseCase(portfolioSvc PortfolioSvc) *GetPortfolioRiskUseCase {
	return &GetPortfolioRiskUseCase{
		portfolioSvc: portfolioSvc,
	}
}

func (uc *GetPortfolioRiskUseCase) Execute(ctx context.Context, in domain.GetPortfolioRiskInput) (domain.GetPortfolioRiskOutput, error) {
	result, err := uc.portfolioSvc.GetRisk(ctx, in.PortfolioName)
	if err != nil {
		return domain.GetPortfolioRiskOutput{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	return domain.GetPortfolioRiskOutput(result), nil
}
