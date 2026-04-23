package news

import (
	"context"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type StaticProvider struct {
	items map[string][]domain.NewsHeadline
}

func NewStaticProvider() *StaticProvider {
	return &StaticProvider{
		items: map[string][]domain.NewsHeadline{
			"NVDA": {
				{Title: "NVIDIA remains central to AI infrastructure demand", URL: "http://bbc.co.uk"},
				{Title: "Markets continue watching data center growth expectations", URL: "http://bbc.co.uk"},
			},
			"ASML": {
				{Title: "ASML remains tied to semiconductor capex trends", URL: "http://bbc.co.uk"},
				{Title: "Export controls remain a key watch item", URL: "http://bbc.co.uk"},
			},
			"SAP": {
				{Title: "SAP continues enterprise cloud transition focus", URL: "http://bbc.co.uk"},
			},
			"BTC": {
				{Title: "Why did Bitcoin price (BTC USD) and crypto stocks fall today after Fed chair nominee Kevin Warsh's", URL: "http://bbc.co.uk"},
			},
		},
	}
}

func (p *StaticProvider) GetNews(
	_ context.Context,
	symbol string,
	limit int,
) (domain.TickerNewsReport, error) {
	report := domain.TickerNewsReport{
		Ticker: symbol,
	}

	items := p.items[strings.ToUpper(symbol)]

	if limit <= 0 || limit >= len(items) {
		report.Headlines = append([]domain.NewsHeadline(nil), items...)
		return report, nil
	}

	report.Headlines = append([]domain.NewsHeadline(nil), items[:limit]...)
	return report, nil
}
