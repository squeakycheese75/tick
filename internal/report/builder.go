package report

import (
	"context"

	"github.com/squeakycheese75/tick/internal/domain"
)

type (
	PortfolioSvc interface {
		GetAnalysis(ctx context.Context, portfolioName string) (domain.PortfolioAnalysis, error)
		GetRisk(ctx context.Context, portfolioName string) (domain.PortfolioRisk, error)
	}
	NewsSvc interface {
		GetNews(ctx context.Context, ticker string, newsLimit int) (domain.NewsSummary, error)
	}
	PricingSvc interface {
		GetValuationQuote(ctx context.Context, symbol string, targetCurrency string, instrumentCurrency string, instrumentType string) (domain.ValuationQuote, error)
	}
	PortfolioInsights interface {
		TopHoldings(portfolioAnalysis domain.PortfolioAnalysis, limit int) []domain.AnalyzedPosition
		AttentionSignals(portfolioAnalysis domain.PortfolioAnalysis, portfolioRisk domain.PortfolioRisk) []string
	}
	SnapshotSvc interface {
		SaveAndEnrichDailyReport(ctx context.Context, dailyReport domain.DailyReport, analysis domain.PortfolioAnalysis) (domain.DailyReport, error)
	}
)

type ReportBuilder struct {
	portfolioSvc PortfolioSvc
	pricingSvc   PricingSvc
	newsSvc      NewsSvc
	insights     PortfolioInsights
	snapshotSvc  SnapshotSvc
}

func NewReportBuilder(
	portfolioSvc PortfolioSvc,
	pricingSvc PricingSvc,
	newsSvc NewsSvc,
	insights PortfolioInsights,
	snapshotSvc SnapshotSvc,

) *ReportBuilder {
	return &ReportBuilder{
		portfolioSvc: portfolioSvc,
		newsSvc:      newsSvc,
		insights:     insights,
		pricingSvc:   pricingSvc,
		snapshotSvc:  snapshotSvc,
	}
}
