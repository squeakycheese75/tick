package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type PortfolioRepository interface {
	GetByName(ctx context.Context, name string) (domain.Portfolio, error)
}

type PositionRepository interface {
	ListByPortfolio(ctx context.Context, portfolioName string) ([]domain.Position, error)
}

type PortfolioAnalyzer interface {
	Analyze(ctx context.Context, in analysis.AnalyzePortfolioInput) (analysis.PortfolioAnalysis, error)
}

type RiskAnalyzer interface {
	Analyze(in analysis.PortfolioAnalysis) (analysis.PortfolioRisk, error)
}

type PortfolioService struct {
	portfolios        PortfolioRepository
	positions         PositionRepository
	portfolioAnalyzer PortfolioAnalyzer
	riskAnalyzer      RiskAnalyzer
}

func NewPortfolioService(
	portfolios PortfolioRepository,
	positions PositionRepository,
	portfolioAnalyzer PortfolioAnalyzer,
	riskAnalyzer RiskAnalyzer,
) *PortfolioService {
	return &PortfolioService{
		portfolios:        portfolios,
		positions:         positions,
		portfolioAnalyzer: portfolioAnalyzer,
		riskAnalyzer:      riskAnalyzer,
	}
}

func (s *PortfolioService) GetAnalysis(ctx context.Context, portfolioName string) (analysis.PortfolioAnalysis, error) {
	pf, err := s.portfolios.GetByName(ctx, portfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return analysis.PortfolioAnalysis{}, fmt.Errorf(
				"portfolio %q not found. Create it with:\n  tick portfolio create %s --base-currency EUR",
				portfolioName,
				portfolioName,
			)
		}
		return analysis.PortfolioAnalysis{}, fmt.Errorf("get portfolio: %w", err)
	}

	positions, err := s.positions.ListByPortfolio(ctx, portfolioName)
	if err != nil {
		return analysis.PortfolioAnalysis{}, fmt.Errorf("list positions: %w", err)
	}

	result, err := s.portfolioAnalyzer.Analyze(ctx, analysis.AnalyzePortfolioInput{
		Portfolio: pf,
		Positions: positions,
	})
	if err != nil {
		return analysis.PortfolioAnalysis{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	return result, nil
}

func (s *PortfolioService) GetRisk(ctx context.Context, portfolioName string) (analysis.PortfolioRisk, error) {
	portfolioAnalysis, err := s.GetAnalysis(ctx, portfolioName)
	if err != nil {
		return analysis.PortfolioRisk{}, err
	}

	portfolioRisk, err := s.riskAnalyzer.Analyze(portfolioAnalysis)
	if err != nil {
		return analysis.PortfolioRisk{}, fmt.Errorf("analyze risk: %w", err)
	}

	return portfolioRisk, nil
}
