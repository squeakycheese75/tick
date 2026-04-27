package report

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type BuildDailyReportParams struct {
	PortfolioName string
	NewsLimit     int
	SaveSnapshot  bool
}

func (s *ReportBuilder) BuildDailyReport(
	ctx context.Context,
	in BuildDailyReportParams,
) (domain.DailyReport, error) {
	analysis, err := s.portfolioSvc.GetAnalysis(ctx, in.PortfolioName)
	if err != nil {
		return domain.DailyReport{}, fmt.Errorf("get portfolio analysis: %w", err)
	}

	risk, err := s.portfolioSvc.GetRisk(ctx, in.PortfolioName)
	if err != nil {
		return domain.DailyReport{}, fmt.Errorf("get portfolio risk: %w", err)
	}

	report := s.buildDailyReportFromAnalysis(analysis, risk)

	news, err := s.getNewsSummaries(ctx, report.TopHoldings, in.NewsLimit)
	if err != nil {
		return domain.DailyReport{}, err
	}
	report.News = news

	if !in.SaveSnapshot {
		return report, nil
	}

	report, err = s.snapshotSvc.SaveAndEnrichDailyReport(ctx, report, analysis)
	if err != nil {
		return domain.DailyReport{}, fmt.Errorf("save and enrich daily report snapshot: %w", err)
	}

	return report, nil
}

func (s *ReportBuilder) buildDailyReportFromAnalysis(
	analysis domain.PortfolioAnalysis,
	risk domain.PortfolioRisk,
) domain.DailyReport {
	topPositions := s.insights.TopHoldings(analysis, 3)

	return domain.DailyReport{
		Portfolio:   assemblePortfolioSummary(analysis),
		TopHoldings: assembleHoldingSummary(topPositions),
		Risk:        assembleRiskSummary(risk),
		Attention:   s.insights.AttentionSignals(analysis, risk),
	}
}
