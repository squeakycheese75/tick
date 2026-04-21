package report

import (
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type TopHoldingReport struct {
	Symbol          string
	Weight          float64
	MarketValueBase float64
	QuotedPrice     float64
	PriceCurrency   string
	ChangePercent   float64
}

type RiskReport struct {
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type TickerNewsReport struct {
	Ticker    string
	Headlines []domain.NewsHeadline
	Summary   string
}

type DailyReport struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64

	TopHoldings []TopHoldingReport
	Risk        RiskReport
	News        []TickerNewsReport
	Attention   []string
}

type DailyReportResult struct {
	Report   DailyReport
	Analysis analysis.PortfolioAnalysis
	Risk     analysis.PortfolioRisk
}
