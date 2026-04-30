package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

type CoinGeckoProvider struct {
	httpClient *http.Client
	baseURL    string
}

func NewCoinGeckoProvider() *CoinGeckoProvider {
	return &CoinGeckoProvider{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    "https://api.coingecko.com/api/v3",
	}
}

func (p *CoinGeckoProvider) GetQuote(ctx context.Context, ticker string) (domain.Quote, error) {
	coinID, err := mapCryptoTickerToCoinGeckoID(ticker)
	if err != nil {
		return domain.Quote{}, err
	}

	u, err := url.Parse(p.baseURL + "/simple/price")
	if err != nil {
		return domain.Quote{}, fmt.Errorf("build coingecko url: %w", err)
	}

	q := u.Query()
	q.Set("ids", coinID)
	q.Set("vs_currencies", "usd")
	q.Set("include_24hr_change", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("coingecko request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return domain.Quote{}, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	var payload map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.Quote{}, fmt.Errorf("decode coingecko response: %w", err)
	}

	coinData, ok := payload[coinID]
	if !ok {
		return domain.Quote{}, fmt.Errorf("coingecko returned no data for %q", coinID)
	}

	price, ok := coinData["usd"]
	if !ok {
		return domain.Quote{}, fmt.Errorf("coingecko returned no usd price for %q", coinID)
	}

	changePctPercent := coinData["usd_24h_change"]
	changePctRatio := changePctPercent / 100

	previous := price / (1 + changePctRatio)
	change := price - previous

	return domain.Quote{
		Symbol:        strings.ToUpper(strings.TrimSpace(ticker)),
		Price:         price,
		PriceCurrency: "USD",
		PreviousClose: previous,
		Change:        change,
		ChangePercent: changePctPercent,
		Source:        "coingecko",
	}, nil
}

func mapCryptoTickerToCoinGeckoID(ticker string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(ticker)) {
	case "BTC":
		return "bitcoin", nil
	case "ETH":
		return "ethereum", nil
	default:
		return "", fmt.Errorf("unsupported crypto ticker %q", ticker)
	}
}
