package usecase

type GetPortfolioSummaryUsecaseInput struct {
	PortfolioName string
}

type GetPortfolioSummaryUsecaseOutput struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	Positions     []SummaryPosition
}

type SummaryPosition struct {
	Ticker             string
	BaseCurrency       string
	InstrumentCurrency string
	AvgCost            float64
	ConvertedPrice     float64
	QuotedPrice        float64
	Weight             float64
	MarketValueBase    float64
	FXRate             float64
	Quantity           float64
}

type CreatePortfolioUsecaseInput struct {
	PortfolioName string
	BaseCurrency  string
}

type CreatePortfolioUsecaseOutout struct {
	PortfolioName string
	BaseCurrency  string
}

type AddPositionToPortfolioUseCaseInput struct {
	Ticker        string
	Qty           float64
	Currency      string
	AvgCost       float64
	PortfolioName string
}

type AddPositionToPortfolioUseCaseOutput struct {
	Ticker        string
	Qty           float64
	Currency      string
	AvgCost       float64
	PortfolioName string
}

type GetPortfolioRiskInput struct {
	PortfolioName string
}

type GetPortfolioRiskOutput struct {
	PortfolioName     string
	BaseCurrency      string
	PositionCount     int
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type GetDailyBriefInput struct {
	PortfolioName string
	NewsLimit     int
}

type GetDailyBriefOutput struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	TopHoldings   []DailyHolding
	Risk          DailyRisk
	News          []DailyNews
	Attention     []string
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

type NewsHeadline struct {
	Title string
	URL   string
}
