package domain

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
