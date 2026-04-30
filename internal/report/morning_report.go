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
	analysis, err := s.analysisSvc.GetAnalysis(ctx, in.PortfolioName)
	if err != nil {
		return domain.BriefReport{}, fmt.Errorf("get portfolio analysis: %w", err)
	}

	out := domain.BriefReport{
		Greeting:  assembleGreeting(),
		Portfolio: assemblePortfolioSummary(analysis),
		Movers:    assembleHoldingSummary(analysis.AnalyzedPositions),
	}

	news, err := s.getNewsSummaries(ctx, out.Movers, 1)
	if err != nil {
		return domain.BriefReport{}, fmt.Errorf("get news summaries: %w", err)
	}

	out.News = news

	sortHoldingsByAbsValueChange(out.Movers.Holdings)

	return out, nil
}
