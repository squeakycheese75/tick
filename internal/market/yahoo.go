package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

type YahooPriceProvider struct {
	httpClient *http.Client
	baseURL    string
}

func NewYahooPriceProvider(httpClient *http.Client) *YahooPriceProvider {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &YahooPriceProvider{
		httpClient: httpClient,
		baseURL:    "https://query1.finance.yahoo.com",
	}
}

func (p *YahooPriceProvider) GetQuote(ctx context.Context, in GetQuoteParams) (domain.Quote, error) {
	endpoint := fmt.Sprintf(
		"%s/v8/finance/chart/%s?interval=1d&range=1d",
		p.baseURL,
		in.ProviderSymbol,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.Quote{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")
	req.Header.Set("Accept", "application/json,text/plain,*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return domain.Quote{}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return domain.Quote{}, fmt.Errorf("yahoo quote request failed for %s: status %d", in.ProviderSymbol, res.StatusCode)
	}

	var body yahooChartResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return domain.Quote{}, err
	}

	return parseYahooQuote(in.ProviderSymbol, body)
}

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol             string  `json:"symbol"`
				Currency           string  `json:"currency"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
			} `json:"meta"`
		} `json:"result"`
		Error any `json:"error"`
	} `json:"chart"`
}

func parseYahooQuote(requestedSymbol string, body yahooChartResponse) (domain.Quote, error) {
	if len(body.Chart.Result) == 0 {
		return domain.Quote{}, fmt.Errorf("no yahoo quote result for %s", requestedSymbol)
	}

	meta := body.Chart.Result[0].Meta

	if meta.RegularMarketPrice == 0 {
		return domain.Quote{}, fmt.Errorf("missing yahoo price for %s", requestedSymbol)
	}

	symbol := meta.Symbol
	if symbol == "" {
		symbol = requestedSymbol
	}

	previousClose := meta.ChartPreviousClose
	change := meta.RegularMarketPrice - previousClose

	changePercent := 0.0
	if previousClose != 0 {
		changePercent = (change / previousClose) * 100
	}

	return domain.Quote{
		Symbol:        symbol,
		Price:         meta.RegularMarketPrice,
		PreviousClose: previousClose,
		Change:        change,
		ChangePercent: changePercent,
		PriceCurrency: meta.Currency,
		Source:        "yahoo",
	}, nil
}
