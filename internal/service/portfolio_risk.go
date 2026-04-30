package service

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type PortfolioRiskSvc struct {
	portfolios   PortfolioRepository
	positions    PositionRepository
	riskAnalyzer RiskAnalyzer
}

func NewPortfolioRiskSvc(
	portfolios PortfolioRepository,
	positions PositionRepository,
	riskAnalyzer RiskAnalyzer,
) *PortfolioRiskSvc {
	return &PortfolioRiskSvc{
		portfolios:   portfolios,
		positions:    positions,
		riskAnalyzer: riskAnalyzer,
	}
}

func (s *PortfolioRiskSvc) GetRisk(ctx context.Context, portfolioAnalysis domain.PortfolioAnalysis) (domain.PortfolioRisk, error) {
	portfolioRisk, err := s.riskAnalyzer.Analyze(portfolioAnalysis)
	if err != nil {
		return domain.PortfolioRisk{}, fmt.Errorf("analyze risk: %w", err)
	}

	return domain.PortfolioRisk{
		PortfolioName:     portfolioRisk.PortfolioName,
		BaseCurrency:      portfolioRisk.BaseCurrency,
		LargestPosition:   portfolioRisk.LargestPosition,
		PositionCount:     portfolioRisk.PositionCount,
		LargestWeight:     portfolioRisk.LargestWeight,
		Top3Concentration: portfolioRisk.Top3Concentration,
		Observations:      portfolioRisk.Observations,
	}, nil
}
