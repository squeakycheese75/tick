package domain

import (
	"time"
)

type PortfolioSnapshot struct {
	ID            int64
	PortfolioName string
	BaseCurrency  string
	CapturedAt    time.Time
	TotalValue    float64
	Positions     []PortfolioSnapshotPosition
}

type PortfolioSnapshotPosition struct {
	Symbol             string
	Quantity           float64
	InstrumentCurrency string
	QuotedPrice        float64
	FXRate             float64
	MarketValueBase    float64
	Weight             float64
}

type TickerNews struct {
	Ticker    string
	Headlines []NewsHeadline
}

type DailyHolding struct {
	Ticker          string
	Weight          float64
	MarketValueBase float64
	QuotedPrice     float64
	PriceCurrency   string
	ChangePercent   float64
}

type DailyRisk struct {
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type DailyNews struct {
	Ticker    string
	Headlines []NewsHeadline
}

type DailyReport struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64

	ChangeSinceLastSnapshot *ValueChangeReport

	TopHoldings []TopHoldingReport
	Risk        RiskReport
	News        []TickerNewsReport
	Attention   []string
}

type TickerNewsReport struct {
	Ticker    string
	Headlines []NewsHeadline
	Summary   string
}

type RiskReport struct {
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type TopHoldingReport struct {
	Symbol          string
	Weight          float64
	MarketValueBase float64
	QuotedPrice     float64
	PriceCurrency   string
	ChangePercent   float64

	SinceLastSnapshot *HoldingSnapshotChangeReport
}

type ValueChangeReport struct {
	Absolute float64
	Percent  float64
}

type HoldingSnapshotChangeReport struct {
	ValueAbsolute float64
	ValuePercent  float64
}

type DailyReportResult struct {
	Report   DailyReport
	Analysis PortfolioAnalysis
	Risk     PortfolioRisk
}

type AnalyzedPosition struct {
	Symbol             string
	Quantity           float64
	AvgCost            float64
	InstrumentCurrency string
	QuotedPrice        float64
	QuotedChange       float64
	QuotedChangePct    float64
	PriceCurrency      string
	FXRate             float64
	ConvertedPrice     float64
	ConvertedChange    float64
	MarketValueBase    float64
	Weight             float64
}
