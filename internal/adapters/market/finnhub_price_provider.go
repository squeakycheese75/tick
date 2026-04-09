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
	C float64 `json:"c"` // current price
}

func (p *FinnhubPriceProvider) GetQuote(ctx context.Context, ticker string) (domain.Quote, error) {
	url := fmt.Sprintf(
		"https://finnhub.io/api/v1/quote?symbol=%s&token=%s",
		ticker,
		p.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return domain.Quote{Price: 0}, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return domain.Quote{Price: 0}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var data finnhubQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return domain.Quote{Price: 0}, err
	}

	if data.C == 0 {
		return domain.Quote{Price: 0}, fmt.Errorf("no price returned for %s", ticker)
	}

	// Finnhub does not always return currency → assume USD for now
	return domain.Quote{
		Ticker:        ticker,
		Price:         data.C,
		PriceCurrency: "USD",
		Source:        "finnhub",
	}, nil
}
