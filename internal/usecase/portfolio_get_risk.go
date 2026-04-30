package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type GetPortfolioRiskUseCase struct {
	anaysisSvc AnaysisSvc
	riskSvc    RiskSvc
}

func NewGetPortfolioRiskUseCase(anaysisSvc AnaysisSvc, riskSvc RiskSvc) *GetPortfolioRiskUseCase {
	return &GetPortfolioRiskUseCase{
		anaysisSvc: anaysisSvc,
		riskSvc:    riskSvc,
	}
}

func (uc *GetPortfolioRiskUseCase) Execute(ctx context.Context, in domain.GetPortfolioRiskInput) (domain.GetPortfolioRiskOutput, error) {
	analysis, err := uc.anaysisSvc.GetAnalysis(ctx, in.PortfolioName)
	if err != nil {
		return domain.GetPortfolioRiskOutput{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	result, err := uc.riskSvc.GetRisk(ctx, analysis)
	if err != nil {
		return domain.GetPortfolioRiskOutput{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	return domain.GetPortfolioRiskOutput(result), nil
}
