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
				{Title: "NVIDIA remains central to AI infrastructure demand"},
				{Title: "Markets continue watching data center growth expectations"},
			},
			"ASML": {
				{Title: "ASML remains tied to semiconductor capex trends"},
				{Title: "Export controls remain a key watch item"},
			},
			"SAP": {
				{Title: "SAP continues enterprise cloud transition focus"},
			},
		},
	}
}

func (p *StaticProvider) GetNews(
	_ context.Context,
	ticker string,
	limit int,
) ([]domain.NewsHeadline, error) {
	items := p.items[strings.ToUpper(ticker)]

	if limit <= 0 || limit >= len(items) {
		return append([]domain.NewsHeadline(nil), items...), nil
	}

	return append([]domain.NewsHeadline(nil), items[:limit]...), nil
}
