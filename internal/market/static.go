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

func (p *StaticCryptoPriceProvider) GetQuote(_ context.Context, in GetQuoteParams) (domain.Quote, error) {
	v, ok := p.prices[strings.ToUpper(in.Symbol)]
	if !ok {
		return domain.Quote{}, fmt.Errorf("price not found for %s", in.Symbol)
	}

	change := 0.0
	changePct := 0.0

	if v.PreviousClose > 0 {
		change = v.Price - v.PreviousClose
		changePct = (change / v.PreviousClose) * 100
	}

	return domain.Quote{
		Symbol:        in.Symbol,
		Price:         v.Price,
		PriceCurrency: v.Currency,
		PreviousClose: v.PreviousClose,
		Change:        change,
		ChangePercent: changePct,
		Source:        "static",
	}, nil
}

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

func (p *StaticPriceProvider) GetQuote(_ context.Context, in GetQuoteParams) (domain.Quote, error) {
	v, ok := p.prices[strings.ToUpper(in.Symbol)]
	if !ok {
		return domain.Quote{}, fmt.Errorf("price not found for %s", in.Symbol)
	}

	change := 0.0
	changePct := 0.0

	if v.PreviousClose > 0 {
		change = v.Price - v.PreviousClose
		changePct = (change / v.PreviousClose) * 100
	}

	return domain.Quote{
		Symbol:        in.Symbol,
		Price:         v.Price,
		PriceCurrency: v.Currency,
		PreviousClose: v.PreviousClose,
		Change:        change,
		ChangePercent: changePct,
		Source:        "static",
	}, nil
}

type StaticFXProvider struct {
	rates map[string]float64
}

func NewStaticFXProvider() *StaticFXProvider {
	return &StaticFXProvider{
		rates: map[string]float64{
			"EUR:EUR": 1.0,
			"USD:USD": 1.0,
			"GBP:GBP": 1.0,
			"USD:EUR": 0.92,
			"EUR:USD": 1.09,
			"GBP:EUR": 1.17,
			"EUR:GBP": 0.85,
			"USD:GBP": 0.78,
			"GBP:USD": 1.28,
		},
	}
}

func (f *StaticFXProvider) GetRate(_ context.Context, from string, to string) (domain.FXRate, error) {
	key := strings.ToUpper(from) + ":" + strings.ToUpper(to)
	rate, ok := f.rates[key]
	if !ok {
		return domain.FXRate{}, fmt.Errorf("fx rate not found for %s", key)
	}
	return domain.FXRate{
		Rate:          rate,
		BaseCurrency:  from,
		QuoteCurrency: to,
		Source:        "static",
	}, nil
}

type StaticCommodityPriceProvider struct {
	prices map[string]domain.Quote
}

func NewStaticCommodityPriceProvider() *StaticCommodityPriceProvider {
	return &StaticCommodityPriceProvider{
		prices: map[string]domain.Quote{
			"GOLD": {
				Symbol:        "GOLD",
				Price:         3400.00,
				PreviousClose: 3380.00,
				Change:        20.00,
				PriceCurrency: "USD",
				Source:        "static",
			},
			"SILVER": {
				Symbol:        "SILVER",
				Price:         39.00,
				PreviousClose: 38.50,
				Change:        0.50,
				PriceCurrency: "USD",
				Source:        "static",
			},
		},
	}
}

func (p *StaticCommodityPriceProvider) GetQuote(
	ctx context.Context,
	in GetQuoteParams,
) (domain.Quote, error) {
	quote, ok := p.prices[in.Symbol]
	if !ok {
		return domain.Quote{}, fmt.Errorf(
			"static commodity quote not found for %s",
			in.Symbol,
		)
	}

	return quote, nil
}
