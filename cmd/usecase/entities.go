package usecase

const (
	basePortfolioCcy = "EUR"
)

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
	Quantity           float64
	InstrumentCurrency string
	BaseCurrency       string
	AvgCost            float64
	CurrentPrice       float64
	FXRate             float64
	MarketValueBase    float64
	Weight             float64
}
type CreatePortfolioUsecaseInput struct {
	Name         string
	BaseCurrency string
}

type CreatePortfolioUsecaseOutout struct {
	Name         string
	BaseCurrency string
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
