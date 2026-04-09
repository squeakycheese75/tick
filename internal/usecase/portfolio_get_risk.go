package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type GetPortfolioRiskUseCase struct {
	portfolios        PortfolioRepository
	positions         PositionRepository
	portfolioAnalyzer PortfolioAnalyzer
	riskAnalyzer      RiskAnalyzer
}

func NewGetPortfolioRiskUseCase(portfolioRepo PortfolioRepository, positionRepo PositionRepository, portfolioAnalyzer PortfolioAnalyzer, riskAnalyzer RiskAnalyzer) *GetPortfolioRiskUseCase {
	return &GetPortfolioRiskUseCase{
		portfolios:        portfolioRepo,
		positions:         positionRepo,
		portfolioAnalyzer: portfolioAnalyzer,
		riskAnalyzer:      riskAnalyzer,
	}
}

func (uc *GetPortfolioRiskUseCase) Execute(ctx context.Context, in GetPortfolioRiskInput) (GetPortfolioRiskOutput, error) {
	pf, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return GetPortfolioRiskOutput{}, fmt.Errorf(
				"portfolio %q not found. Create it with:\n  tick portfolio create %s --base-currency EUR",
				in.PortfolioName,
				in.PortfolioName,
			)
		}
		return GetPortfolioRiskOutput{}, fmt.Errorf("get portfolio: %w", err)
	}

	positions, err := uc.positions.ListByPortfolio(ctx, in.PortfolioName)
	if err != nil {
		return GetPortfolioRiskOutput{}, fmt.Errorf("list positions: %w", err)
	}

	result, err := uc.portfolioAnalyzer.Analyze(ctx, analysis.AnalyzePortfolioInput{
		Portfolio: pf,
		Positions: positions,
	})
	if err != nil {
		return GetPortfolioRiskOutput{}, fmt.Errorf("portfolio analyzer: %w", err)
	}

	analyzedResult, err := uc.riskAnalyzer.Analyze(result)
	if err != nil {
		return GetPortfolioRiskOutput{}, fmt.Errorf("risk analyzer: %w", err)
	}

	return mapPortfolioRiskToOutput(analyzedResult), nil
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
