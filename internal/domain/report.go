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

type Holding struct {
	Symbol            string
	Quantity          float64
	Weight            float64
	MarketValueBase   float64
	QuotedPrice       float64
	PriceCurrency     string
	ChangeAbsolute    float64
	ChangePercent     float64
	SinceLastSnapshot *ValueChangeSummary
}

type HoldingSummary struct {
	Holdings   []Holding
	TotalValue float64
	Change     *ValueChangeSummary
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
	Portfolio   PortfolioSummary
	TopHoldings HoldingSummary
	Risk        RiskSummary
	News        []NewsSummary
	Attention   []string
	Targets     []TargetStatus
}

type TargetStatus struct {
	Symbol       string
	Type         TargetType
	CurrentPrice float64
	TargetPrice  float64
	Currency     string
	Hit          bool
	DistancePct  float64
}

type BriefReport struct {
	Greeting  string
	Portfolio PortfolioSummary
	Movers    HoldingSummary
	Markets   []MarketSummary
	News      []NewsSummary
}

type PortfolioSummary struct {
	Name         string
	BaseCurrency string
	TotalValue   float64
	Change       *ValueChangeSummary
}

type MarketSummary struct {
	Symbol        string
	Price         float64
	Change        float64
	ChangePercent float64
	Currency      string
}

type NewsSummary struct {
	Ticker    string
	Headlines []NewsHeadline
	Summary   string
}

type RiskSummary struct {
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type AttentionSummary struct {
	Signal string
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

type ValueChangeSummary struct {
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
