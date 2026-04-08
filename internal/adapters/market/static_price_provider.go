package market

import (
	"context"
	"fmt"
	"strings"
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

func (p *StaticPriceProvider) GetPrice(_ context.Context, ticker string) (float64, string, error) {
	v, ok := p.prices[strings.ToUpper(ticker)]
	if !ok {
		return 0, "", fmt.Errorf("price not found for %s", ticker)
	}
	return v.Price, v.Currency, nil
}
