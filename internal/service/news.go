package service

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/report"
)

type NewsProvider interface {
	GetNews(ctx context.Context, ticker string, limit int) ([]domain.NewsHeadline, error)
}

type NewsService struct {
	provider NewsProvider
}

func NewNewsService(provider NewsProvider) *NewsService {
	return &NewsService{
		provider: provider,
	}
}

func (s *NewsService) GetNews(
	ctx context.Context,
	ticker string,
	newsLimit int,
) (report.TickerNewsReport, error) {
	headlines, err := s.provider.GetNews(ctx, ticker, newsLimit)
	if err != nil {
		return report.TickerNewsReport{}, fmt.Errorf("get news for %s: %w", ticker, err)
	}

	return report.TickerNewsReport{
		Ticker:    ticker,
		Headlines: headlines,
	}, nil
}
