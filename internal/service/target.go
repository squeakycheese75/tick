package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type TargetSvc struct {
	portfolios PortfolioRepository
	targets    TargetRepository
}

func NewTargetSvc(
	portfolios PortfolioRepository,
	targets TargetRepository,
) *TargetSvc {
	return &TargetSvc{
		portfolios: portfolios,
		targets:    targets,
	}
}

func (s *TargetSvc) EvaluateTargets(
	ctx context.Context,
	portfolioName string,
	analysis domain.PortfolioAnalysis,
) ([]domain.TargetStatus, error) {
	portfolio, err := s.portfolios.GetByName(ctx, portfolioName)
	if err != nil {
		return nil, fmt.Errorf("get portfolio: %w", err)
	}

	targets, err := s.targets.ListByPortfolioID(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("list targets: %w", err)
	}

	quotes := quotesBySymbol(analysis)

	statuses := make([]domain.TargetStatus, 0, len(targets))

	for _, target := range targets {
		quote, ok := quotes[strings.ToUpper(target.Symbol)]
		if !ok {
			continue
		}

		statuses = append(statuses, evaluateTarget(target, quote))
	}

	return statuses, nil
}

func evaluateTarget(target domain.Target, quote domain.Quote) domain.TargetStatus {
	distancePct := ((target.TargetPrice - quote.Price) / quote.Price) * 100

	hit := false

	if target.Type == domain.TargetTypeTakeProfit {
		hit = quote.Price >= target.TargetPrice
	}

	if target.Type == domain.TargetTypeStopLoss {
		hit = quote.Price <= target.TargetPrice
	}

	return domain.TargetStatus{
		Symbol:       target.Symbol,
		Type:         target.Type,
		CurrentPrice: quote.Price,
		TargetPrice:  target.TargetPrice,
		Currency:     target.QuoteCurrency,
		Hit:          hit,
		DistancePct:  distancePct,
	}
}

func quotesBySymbol(analysis domain.PortfolioAnalysis) map[string]domain.Quote {
	out := make(map[string]domain.Quote)

	for _, holding := range analysis.AnalyzedPositions {
		out[strings.ToUpper(holding.Symbol)] = domain.Quote{
			Symbol:        holding.Symbol,
			Price:         holding.QuotedPrice,
			PriceCurrency: holding.PriceCurrency,
		}
	}

	return out
}
