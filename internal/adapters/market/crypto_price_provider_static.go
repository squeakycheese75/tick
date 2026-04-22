package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type StaticCryptoPriceProvider struct {
	prices map[string]struct {
		Price         float64
		PreviousClose float64
		Currency      string
	}
}

func NewStaticCryptoPriceProvider() *StaticCryptoPriceProvider {
	return &StaticCryptoPriceProvider{
		prices: map[string]struct {
			Price         float64
			PreviousClose float64
			Currency      string
		}{
			"BTC": {Price: 78260.452, Currency: "USD"},
			"ETH": {Price: 2394.62, Currency: "USD"},
		},
	}
}

func (p *StaticCryptoPriceProvider) GetQuote(_ context.Context, ticker string) (domain.Quote, error) {
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
		Ticker:        ticker,
		Price:         v.Price,
		PriceCurrency: v.Currency,
		PreviousClose: v.PreviousClose,
		Change:        change,
		ChangePercent: changePct,
		Source:        "static",
	}, nil
}
