package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
}

func (p *FinnhubPriceProvider) GetQuote(ctx context.Context, ticker string) (domain.Quote, error) {
	url := fmt.Sprintf(
		"https://finnhub.io/api/v1/quote?symbol=%s&token=%s",
		ticker,
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
		return domain.Quote{}, fmt.Errorf("no price returned for %s", ticker)
	}

	change := 0.0
	changePercent := 0.0
	if data.PC > 0 {
		change = data.C - data.PC
		changePercent = (change / data.PC) * 100
	}

	return domain.Quote{
		Ticker:        ticker,
		Price:         data.C,
		PriceCurrency: "USD",
		PreviousClose: data.PC,
		Change:        change,
		ChangePercent: changePercent,
		Source:        "finnhub",
	}, nil
}
