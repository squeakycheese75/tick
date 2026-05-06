package domain

import "time"

type Quote struct {
	Symbol        string
	Price         float64
	PriceCurrency string
	PreviousClose float64
	Change        float64
	ChangePercent float64
	Source        string
	Stale         bool
	AsOf          time.Time
}

type FXRate struct {
	BaseCurrency  string
	QuoteCurrency string
	Rate          float64
	Source        string
}

type ValuationQuote struct {
	Quote                  Quote
	TargetCurrency         string
	FXRate                 float64
	ConvertedPrice         float64
	ConvertedPreviousClose float64
	ConvertedChange        float64
}
