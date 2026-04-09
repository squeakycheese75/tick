package domain

type PortfolioAnalysis struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	Positions     []Position
}
