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
	// AvgCostBase is the average cost per unit in the portfolio base currency at purchase; 0 when unknown.
	AvgCostBase float64
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
	// AssetPnL is the part of UnrealizedPnL from the price moving, valued at today's FX rate.
	AssetPnL float64
	// FXPnL is the part of UnrealizedPnL from the FX rate moving since purchase. Only meaningful when FXKnown.
	FXPnL float64
	// FXKnown is false when the purchase FX rate was not recorded, so no split is possible.
	FXKnown          bool
	UnrealizedPnLPct float64
}

type Summary struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	Positions     []SummaryPosition
}
type InstrumentType string

type Instrument struct {
	Symbol         string
	ProviderSymbol string
	InstrumentType InstrumentType
	QuoteCurrency  string
	Exchange       string
}

type ProviderSymbol struct {
	Provider string
	Symbol   string
}

const (
	InstrumentTypeEquity    InstrumentType = "equity"
	InstrumentTypeCrypto    InstrumentType = "crypto"
	InstrumentTypeFX        InstrumentType = "fx"
	InstrumentTypeETF       InstrumentType = "etf"
	InstrumentTypeBond      InstrumentType = "bond"
	InstrumentTypeCommodity InstrumentType = "commodity"
	InstrumentTypeFund      InstrumentType = "fund"
)

// PortfolioListItem is one row of the portfolio list.
type PortfolioListItem struct {
	Name          string
	BaseCurrency  string
	PositionCount int
}

// ListPortfoliosOutput is every portfolio, ordered by name.
type ListPortfoliosOutput struct {
	Portfolios []PortfolioListItem
}
