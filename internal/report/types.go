package report

import "github.com/squeakycheese75/tick/internal/domain"

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
