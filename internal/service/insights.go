package service

import (
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type InsightsSvc struct{}

func NewInsightsSvc() *InsightsSvc {
	return &InsightsSvc{}
}

func (b *InsightsSvc) TopHoldings(
	portfolioAnalysis domain.PortfolioAnalysis,
	limit int,
) []domain.AnalyzedPosition {
	if limit <= 0 {
		return nil
	}

	if len(portfolioAnalysis.AnalyzedPositions) < limit {
		limit = len(portfolioAnalysis.AnalyzedPositions)
	}

	result := make([]domain.AnalyzedPosition, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, portfolioAnalysis.AnalyzedPositions[i])
	}

	return result
}

func (b *InsightsSvc) AttentionSignals(
	portfolioAnalysis domain.PortfolioAnalysis,
	portfolioRisk domain.PortfolioRisk,
) []string {
	attention := make([]string, 0)

	if len(portfolioAnalysis.AnalyzedPositions) == 0 {
		return append(attention, "Portfolio is empty")
	}

	if portfolioRisk.LargestWeight >= 0.20 {
		attention = append(
			attention,
			fmt.Sprintf(
				"%s is %.2f%% of the portfolio",
				portfolioRisk.LargestPosition,
				portfolioRisk.LargestWeight*100,
			),
		)
	}

	if portfolioRisk.Top3Concentration >= 0.60 {
		attention = append(
			attention,
			fmt.Sprintf(
				"Top 3 positions are %.2f%% of the portfolio",
				portfolioRisk.Top3Concentration*100,
			),
		)
	}

	if len(attention) == 0 {
		attention = append(attention, "No major portfolio concentration issues detected")
	}

	return attention
}
