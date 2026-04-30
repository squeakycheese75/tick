package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/analysis"
	"github.com/squeakycheese75/tick/internal/domain"
)

type PortfolioAnalysisSvc struct {
	portfolios        PortfolioRepository
	positions         PositionRepository
	portfolioAnalyzer PortfolioAnalyzer
	// riskAnalyzer      RiskAnalyzer
}

func NewPortfolioAnalysisSvc(
	portfolios PortfolioRepository,
	positions PositionRepository,
	portfolioAnalyzer PortfolioAnalyzer,
	// riskAnalyzer RiskAnalyzer,
) *PortfolioAnalysisSvc {
	return &PortfolioAnalysisSvc{
		portfolios:        portfolios,
		positions:         positions,
		portfolioAnalyzer: portfolioAnalyzer,
		// riskAnalyzer:      riskAnalyzer,
	}
}

func (s *PortfolioAnalysisSvc) GetAnalysis(ctx context.Context, portfolioName string) (domain.PortfolioAnalysis, error) {
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
				InstrumentType: domain.InstrumentType(p.Instrument.InstrumentType),
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
