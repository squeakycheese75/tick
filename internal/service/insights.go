package service

import (
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type PortfolioInsights struct{}

func NewPortfolioInsights() *PortfolioInsights {
	return &PortfolioInsights{}
}

func (b *PortfolioInsights) TopHoldings(
	portfolioAnalysis analysis.PortfolioAnalysis,
	limit int,
) []analysis.AnalyzedPosition {
	if limit <= 0 {
		return nil
	}

	if len(portfolioAnalysis.AnalyzedPositions) < limit {
		limit = len(portfolioAnalysis.AnalyzedPositions)
	}

	result := make([]analysis.AnalyzedPosition, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, portfolioAnalysis.AnalyzedPositions[i])
	}

	return result
}

func (b *PortfolioInsights) AttentionSignals(
	portfolioAnalysis analysis.PortfolioAnalysis,
	portfolioRisk analysis.PortfolioRisk,
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
