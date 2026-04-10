package report

type DailyReport struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64

	TopHoldings []TopHoldingReport
	Risk        RiskReport
	News        []TickerNewsReport
	Attention   []string
}
