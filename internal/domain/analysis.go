package domain

import "time"

type PortfolioRisk struct {
	PortfolioName     string
	BaseCurrency      string
	PositionCount     int
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type PortfolioAnalysis struct {
	PortfolioName     string
	BaseCurrency      string
	AnalyzedPositions []AnalyzedPosition
	TotalValue        float64
	ValuationIssues   []ValuationIssue
}

type ValuationIssueType string

const (
	ValuationIssueMissingPrice ValuationIssueType = "missing_price"
	ValuationIssueMissingFX    ValuationIssueType = "missing_fx"
	ValuationIssueProvider     ValuationIssueType = "provider_error"
)

type ValuationIssue struct {
	Symbol         string
	InstrumentType string
	Quantity       float64
	Type           ValuationIssueType
	Message        string
	Hint           string
}

// PricesAsOf is the oldest quote time behind the valuation, or zero when no provider reported one.
func (a PortfolioAnalysis) PricesAsOf() time.Time {
	var oldest time.Time

	for _, p := range a.AnalyzedPositions {
		if p.QuoteAsOf.IsZero() {
			continue
		}
		if oldest.IsZero() || p.QuoteAsOf.Before(oldest) {
			oldest = p.QuoteAsOf
		}
	}

	return oldest
}
