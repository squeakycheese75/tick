package domain

type Quote struct {
	Ticker        string
	Price         float64
	PriceCurrency string
	Source        string
}

type ValuationQuote struct {
	Quote          Quote
	TargetCurrency string
	FXRate         float64
	ConvertedPrice float64
}
