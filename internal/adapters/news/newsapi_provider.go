package news

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

const language = "en"

type NewsAPIProvider struct {
	apiKey     string
	httpClient *http.Client
	hints      map[string][]string
}

func NewNewsAPIProvider(apiKey string, keywordHints map[string][]string) *NewsAPIProvider {
	return &NewsAPIProvider{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		hints:      keywordHints,
	}
}

func (p *NewsAPIProvider) GetNews(
	ctx context.Context,
	ticker string,
	limit int,
) (domain.NewsSummary, error) {

	query := buildQuery(ticker)

	u, _ := url.Parse("https://newsapi.org/v2/everything")
	q := u.Query()
	q.Set("q", query)
	q.Set("sortBy", "publishedAt")
	q.Set("pageSize", strconv.Itoa(limit*3)) // fetch extra → filter down
	q.Set("language", language)
	q.Set("apiKey", p.apiKey)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return domain.NewsSummary{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	var raw struct {
		Articles []struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		} `json:"articles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return domain.NewsSummary{}, err
	}

	out := domain.NewsSummary{
		Ticker: ticker,
	}

	for _, a := range raw.Articles {
		if a.Title == "" || a.URL == "" {
			continue
		}

		if isJunkTitle(a.Title) {
			continue
		}

		if !isUsableURL(a.URL) {
			continue
		}

		if !p.isRelevant(ticker, a.Title) {
			continue
		}

		out.Headlines = append(out.Headlines, domain.NewsHeadline{
			Title: a.Title,
			URL:   a.URL,
		})

		if len(out.Headlines) >= limit {
			break
		}
	}

	return out, nil
}

func (p *NewsAPIProvider) isRelevant(ticker, title string) bool {
	t := strings.ToLower(title)
	s := strings.ToLower(ticker)

	if strings.Contains(t, s) {
		return true
	}

	if hints, ok := p.hints[strings.ToUpper(strings.TrimSpace(ticker))]; ok {
		for _, h := range hints {
			if strings.Contains(t, strings.ToLower(h)) {
				return true
			}
		}
	}

	return false
}

func buildQuery(ticker string) string {
	t := strings.ToUpper(strings.TrimSpace(ticker))

	// basic improvement over raw ticker
	return t + " stock"
}

func isUsableURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	host := strings.ToLower(u.Host)
	path := strings.ToLower(u.Path)

	if strings.Contains(host, "consent.yahoo.com") {
		return false
	}

	if strings.Contains(path, "collectconsent") {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

func isJunkTitle(title string) bool {
	t := strings.ToLower(title)

	return strings.Contains(t, "advertisement") ||
		strings.Contains(t, "sponsored") ||
		strings.Contains(t, "promo")
}
