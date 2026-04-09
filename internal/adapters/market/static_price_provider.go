package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type StaticPriceProvider struct {
	prices map[string]struct {
		Price    float64
		Currency string
	}
}

func NewStaticPriceProvider() *StaticPriceProvider {
	return &StaticPriceProvider{
		prices: map[string]struct {
			Price    float64
			Currency string
		}{
			"NVDA": {Price: 400, Currency: "USD"},
			"ASML": {Price: 850, Currency: "EUR"},
			"SAP":  {Price: 180, Currency: "EUR"},
		},
	}
}

func (p *StaticPriceProvider) GetQuote(_ context.Context, ticker string) (domain.Quote, error) {
	v, ok := p.prices[strings.ToUpper(ticker)]
	if !ok {
		return domain.Quote{}, fmt.Errorf("price not found for %s", ticker)
	}
	return domain.Quote{
		Ticker:        ticker,
		Price:         v.Price,
		PriceCurrency: v.Currency,
		Source:        "static",
	}, nil
}
