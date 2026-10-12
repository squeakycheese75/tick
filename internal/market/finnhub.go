package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

type FinnhubPriceProvider struct {
	apiKey string
	client *http.Client
}

func NewFinnhubPriceProvider(apiKey string) *FinnhubPriceProvider {
	return &FinnhubPriceProvider{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

type finnhubQuoteResponse struct {
	C  float64 `json:"c"`  // current price
	PC float64 `json:"pc"` // previous close
	T  int64   `json:"t"`  // unix time of the last trade
}

func (p *FinnhubPriceProvider) GetQuote(ctx context.Context, in GetQuoteParams) (domain.Quote, error) {
	url := fmt.Sprintf(
		"https://finnhub.io/api/v1/quote?symbol=%s&token=%s",
		in.ProviderSymbol,
		p.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return domain.Quote{}, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return domain.Quote{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var data finnhubQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return domain.Quote{}, err
	}

	if data.C == 0 {
		return domain.Quote{}, fmt.Errorf("no price returned for %s", in.ProviderSymbol)
	}

	change := 0.0
	changePercent := 0.0
	if data.PC > 0 {
		change = data.C - data.PC
		changePercent = (change / data.PC) * 100
	}

	var asOf time.Time
	if data.T > 0 {
		asOf = time.Unix(data.T, 0).UTC()
	}

	return domain.Quote{
		Symbol:        in.ProviderSymbol,
		Price:         data.C,
		PriceCurrency: "USD",
		PreviousClose: data.PC,
		Change:        change,
		ChangePercent: changePercent,
		Source:        "finnhub",
		AsOf:          asOf,
	}, nil
}
