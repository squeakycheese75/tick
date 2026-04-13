package repository

import "time"

type Instrument struct {
	ID             int64
	Symbol         string
	ProviderSymbol string
	AssetType      string
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
	Ticker        string
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
