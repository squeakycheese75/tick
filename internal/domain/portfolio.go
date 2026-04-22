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
	Symbol             string
	Quantity           float64
	QuotedPrice        float64
	BaseCurrency       string
	InstrumentCurrency string
	FXRate             float64
	MarketValueBase    float64
	Weight             float64
	AvgCost            float64
	CostBasisBase      float64
	UnrealizedPnL      float64
	UnrealizedPnLPct   float64
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
