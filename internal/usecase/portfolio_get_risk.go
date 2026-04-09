package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type GetPortfolioRiskUseCase struct {
	portfolioSvc PortfolioSvc
}

func NewGetPortfolioRiskUseCase(portfolioSvc PortfolioSvc) *GetPortfolioRiskUseCase {
	return &GetPortfolioRiskUseCase{
		portfolioSvc: portfolioSvc,
	}
}

func (uc *GetPortfolioRiskUseCase) Execute(ctx context.Context, in GetPortfolioRiskInput) (GetPortfolioRiskOutput, error) {
	result, err := uc.portfolioSvc.GetRisk(ctx, in.PortfolioName)
	if err != nil {
		return GetPortfolioRiskOutput{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	return mapPortfolioRiskToOutput(result), nil
}

func mapPortfolioRiskToOutput(a analysis.PortfolioRisk) GetPortfolioRiskOutput {
	return GetPortfolioRiskOutput{
		PortfolioName:     a.PortfolioName,
		BaseCurrency:      a.BaseCurrency,
		PositionCount:     a.PositionCount,
		LargestPosition:   a.LargestPosition,
		LargestWeight:     a.LargestWeight,
		Top3Concentration: a.Top3Concentration,
		Observations:      a.Observations,
	}
}
