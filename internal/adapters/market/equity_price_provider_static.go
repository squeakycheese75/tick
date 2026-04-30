package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type StaticPriceProvider struct {
	prices map[string]struct {
		Price         float64
		PreviousClose float64
		Currency      string
	}
}

func NewStaticPriceProvider() *StaticPriceProvider {
	return &StaticPriceProvider{
		prices: map[string]struct {
			Price         float64
			PreviousClose float64
			Currency      string
		}{
			"NVDA": {Price: 400, PreviousClose: 390, Currency: "USD"},
			"ASML": {Price: 850, PreviousClose: 845, Currency: "EUR"},
			"SAP":  {Price: 180, PreviousClose: 182, Currency: "EUR"},
		},
	}
}

func (p *StaticPriceProvider) GetQuote(_ context.Context, ticker string) (domain.Quote, error) {
	v, ok := p.prices[strings.ToUpper(ticker)]
	if !ok {
		return domain.Quote{}, fmt.Errorf("price not found for %s", ticker)
	}

	change := 0.0
	changePct := 0.0

	if v.PreviousClose > 0 {
		change = v.Price - v.PreviousClose
		changePct = (change / v.PreviousClose) * 100
	}

	return domain.Quote{
		Symbol:        ticker,
		Price:         v.Price,
		PriceCurrency: v.Currency,
		PreviousClose: v.PreviousClose,
		Change:        change,
		ChangePercent: changePct,
		Source:        "static",
	}, nil
}
