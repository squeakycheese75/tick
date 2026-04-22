package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
	"github.com/squeakycheese75/tick/internal/repository"
)

type PortfolioRepository interface {
	GetByName(ctx context.Context, name string) (repository.Portfolio, error)
}

type PositionRepository interface {
	ListByPortfolioID(ctx context.Context, portfolioID int64) ([]repository.Position, error)
}

type PortfolioAnalyzer interface {
	Analyze(ctx context.Context, in analysis.AnalyzePortfolioInput) (domain.PortfolioAnalysis, error)
}

type RiskAnalyzer interface {
	Analyze(in domain.PortfolioAnalysis) (analysis.PortfolioRisk, error)
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

func (s *PortfolioService) GetAnalysis(ctx context.Context, portfolioName string) (domain.PortfolioAnalysis, error) {
	pf, err := s.portfolios.GetByName(ctx, portfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return domain.PortfolioAnalysis{}, fmt.Errorf(
				"portfolio %q not found. Create it with:\n  tick portfolio create %s --base-currency EUR",
				portfolioName,
				portfolioName,
			)
		}
		return domain.PortfolioAnalysis{}, fmt.Errorf("get portfolio: %w", err)
	}

	positions, err := s.positions.ListByPortfolioID(ctx, pf.ID)
	if err != nil {
		return domain.PortfolioAnalysis{}, fmt.Errorf("list positions: %w", err)
	}

	mappedPositions := make([]domain.Position, 0, len(positions))
	for _, p := range positions {
		mappedPositions = append(mappedPositions, domain.Position{
			PortfolioName: pf.Name,
			Instrument: domain.Instrument{
				Symbol:         p.Instrument.Symbol,
				ProviderSymbol: p.Instrument.Symbol,
				AssetType:      p.Instrument.AssetType,
				QuoteCurrency:  p.Instrument.QuoteCurrency,
				Exchange:       p.Instrument.ProviderSymbol,
			},
			Quantity: p.Quantity,
			AvgCost:  p.AvgCost,
		})
	}

	result, err := s.portfolioAnalyzer.Analyze(ctx, analysis.AnalyzePortfolioInput{
		Portfolio: domain.Portfolio{
			Name:         pf.Name,
			BaseCurrency: pf.BaseCurrency,
		},
		Positions: mappedPositions,
	})
	if err != nil {
		return domain.PortfolioAnalysis{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	return result, nil
}

func (s *PortfolioService) GetRisk(ctx context.Context, portfolioName string) (domain.PortfolioRisk, error) {
	portfolioAnalysis, err := s.GetAnalysis(ctx, portfolioName)
	if err != nil {
		return domain.PortfolioRisk{}, err
	}

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
