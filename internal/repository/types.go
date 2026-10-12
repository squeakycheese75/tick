package repository

import "time"

type Instrument struct {
	ID             int64
	Symbol         string
	ProviderSymbol string
	InstrumentType string
	QuoteCurrency  string
	Exchange       string
}

type Portfolio struct {
	ID           int64
	Name         string
	BaseCurrency string
}

// PortfolioOverview is a portfolio with the number of positions in it.
type PortfolioOverview struct {
	Name          string
	BaseCurrency  string
	PositionCount int
}

type Position struct {
	ID         int64
	Instrument Instrument
	Quantity   float64
	AvgCost    float64
	Currency   string
	// AvgCostBase is the average cost per unit in the portfolio's base currency at purchase; 0 when unknown.
	AvgCostBase float64
}

type PortfolioPositions struct {
	PortfolioName string
	Positions     []Position
}

type CreatePositionParams struct {
	InstrumentID int64
	PortfolioID  int64
	Quantity     float64
	AvgCost      float64
	Currency     string
	AvgCostBase  float64
}

type PriceQuote struct {
	Symbol        string
	SourceSymbol  string
	Price         float64
	PriceCurrency string
	PreviousClose float64
	Change        float64
	ChangePercent float64
	Source        string
	// AsOf is when the price was last traded; zero when the provider does not say.
	AsOf time.Time
}

type CachedPriceQuote struct {
	PriceQuote PriceQuote
	FetchedAt  time.Time
}

type CachedFXRate struct {
	BaseCurrency  string
	QuoteCurrency string
	Rate          float64
	Source        string
	FetchedAt     time.Time
}

type Snapshot struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	CapturedAt    time.Time
}

type SnapshotPosition struct {
	Symbol             string
	Quantity           float64
	InstrumentCurrency string
	QuotedPrice        float64
	FXRate             float64
	MarketValueBase    float64
	Weight             float64
}

type PortfolioSnapshot struct {
	ID            int64
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	CapturedAt    time.Time
}

type PortfolioSnapshotPosition struct {
	SnapshotID         int64
	Symbol             string
	Quantity           float64
	InstrumentCurrency string
	QuotedPrice        float64
	FXRate             float64
	MarketValueBase    float64
	Weight             float64
}

type ConsumedPrice struct {
	Symbol   string
	Price    float64
	Currency string
	AsOf     time.Time
	Source   string
}

type Trade struct {
	PortfolioID  int64
	InstrumentID int64
	Key          string
	Side         string
	Quantity     float64
	Price        float64
	Currency     string
	Fee          float64
	FXRate       float64
	TradedAt     time.Time
}
