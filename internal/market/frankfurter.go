package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

type FrankfurterFXProvider struct {
	client  *http.Client
	baseURL string
}

func NewFrankfurterFXProvider() *FrankfurterFXProvider {
	return &FrankfurterFXProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.frankfurter.dev",
	}
}

type frankfurterSingleRateResponse struct {
	Amount float64 `json:"amount"`
	Base   string  `json:"base"`
	Date   string  `json:"date"`
	Rate   float64 `json:"rate"`
}

func (p *FrankfurterFXProvider) GetRate(ctx context.Context, from string, to string) (domain.FXRate, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))

	fxRate := domain.FXRate{
		BaseCurrency:  from,
		QuoteCurrency: to,
		Source:        "frankfurter",
	}

	if from == "" || to == "" {
		return fxRate, fmt.Errorf("from and to currencies are required")
	}

	if from == to {
		fxRate.Rate = 1.0
		return fxRate, nil
	}

	url := fmt.Sprintf("%s/v2/rate/%s/%s", p.baseURL, from, to)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fxRate.Rate = 0
		return fxRate, fmt.Errorf("build frankfurter request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		fxRate.Rate = 0
		return fxRate, fmt.Errorf("perform frankfurter request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		fxRate.Rate = 0
		return fxRate, fmt.Errorf("frankfurter returned status %s", resp.Status)
	}

	var data frankfurterSingleRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fxRate.Rate = 0
		return fxRate, fmt.Errorf("decode frankfurter response: %w", err)
	}

	if data.Rate == 0 {
		fxRate.Rate = 0
		return fxRate, fmt.Errorf("fx rate not found for %s/%s", from, to)
	}

	fxRate.Rate = data.Rate
	return fxRate, nil
}
