package report

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type BuildMorningBriefReportParams struct {
	PortfolioName string
}

func (s *ReportBuilder) BuildMorningBriefReport(ctx context.Context, in BuildMorningBriefReportParams) (domain.BriefReport, error) {
	analysis, err := s.portfolioSvc.GetAnalysis(ctx, in.PortfolioName)
	if err != nil {
		return domain.BriefReport{}, fmt.Errorf("get portfolio analysis: %w", err)
	}

	out := domain.BriefReport{
		Greeting:  assembleGreeting(),
		Portfolio: assemblePortfolioSummary(analysis),
		Movers:    assembleHoldingSummary(analysis.AnalyzedPositions),
	}

	for _, pos := range analysis.AnalyzedPositions {
		newsItems, err := s.newsSvc.GetNews(ctx, pos.Symbol, 1)
		if err != nil {
			continue
		}
		out.News = append(out.News, newsItems)
	}

	if out.Movers.TotalValue != out.Movers.Change.Absolute {
		previousValue := out.Movers.TotalValue - out.Movers.Change.Absolute
		if previousValue > 0 {
			out.Movers.Change.Percent = (out.Movers.Change.Percent / previousValue) * 100
		}
	}

	sortHoldingsByAbsValueChange(out.Movers.Holdings)

	return out, nil
}
