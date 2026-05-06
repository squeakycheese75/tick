package report

import (
	"context"

	"github.com/squeakycheese75/tick/internal/domain"
)

type (
	AnalysisSvc interface {
		GetAnalysis(ctx context.Context, portfolioName string) (domain.PortfolioAnalysis, error)
	}

	RiskSvc interface {
		GetRisk(ctx context.Context, portfolioAnlaysis domain.PortfolioAnalysis) (domain.PortfolioRisk, error)
	}
	NewsSvc interface {
		GetNews(ctx context.Context, ticker string, newsLimit int) (domain.NewsSummary, error)
	}
	PricingSvc interface {
		GetValuationQuote(ctx context.Context, symbol, providerSymbol string, targetCurrency string, instrumentCurrency string, instrumentType string) (domain.ValuationQuote, error)
	}
	PortfolioInsights interface {
		TopHoldings(portfolioAnalysis domain.PortfolioAnalysis, limit int) []domain.AnalyzedPosition
		AttentionSignals(portfolioAnalysis domain.PortfolioAnalysis, portfolioRisk domain.PortfolioRisk) []string
	}
	SnapshotSvc interface {
		SaveAndEnrichDailyReport(ctx context.Context, dailyReport domain.DailyReport, analysis domain.PortfolioAnalysis) (domain.DailyReport, error)
	}
	TargetSvc interface {
		EvaluateTargets(ctx context.Context, portfolioName string, analysis domain.PortfolioAnalysis) ([]domain.TargetStatus, error)
	}
)

type ReportBuilder struct {
	analysisSvc AnalysisSvc
	riskSvc     RiskSvc
	pricingSvc  PricingSvc
	newsSvc     NewsSvc
	insights    PortfolioInsights
	snapshotSvc SnapshotSvc
	targetSvc   TargetSvc
}

func NewReportBuilder(
	analysisSvc AnalysisSvc,
	portfolioRiskSvc RiskSvc,
	pricingSvc PricingSvc,
	newsSvc NewsSvc,
	insights PortfolioInsights,
	snapshotSvc SnapshotSvc,
	targetSvc TargetSvc,

) *ReportBuilder {
	return &ReportBuilder{
		analysisSvc: analysisSvc,
		riskSvc:     portfolioRiskSvc,
		newsSvc:     newsSvc,
		insights:    insights,
		pricingSvc:  pricingSvc,
		snapshotSvc: snapshotSvc,
		targetSvc:   targetSvc,
	}
}
