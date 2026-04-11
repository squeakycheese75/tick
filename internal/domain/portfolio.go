package domain

type Portfolio struct {
	Name         string
	BaseCurrency string
}

type Position struct {
	PortfolioName string
	Instrument    Instrument
	Quantity      float64
	AvgCost       float64
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

type Summary struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	Positions     []SummaryPosition
}

type Instrument struct {
	Symbol         string
	ProviderSymbol string
	AssetType      string
	QuoteCurrency  string
	Exchange       string
}
