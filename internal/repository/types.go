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

type Position struct {
	ID         int64
	Instrument Instrument
	Quantity   float64
	AvgCost    float64
	Currency   string
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
