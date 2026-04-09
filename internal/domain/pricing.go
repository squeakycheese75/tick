package domain

type Quote struct {
	Ticker        string
	Price         float64
	PriceCurrency string
	PreviousClose float64
	Change        float64
	ChangePercent float64
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
