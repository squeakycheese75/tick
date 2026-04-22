package service

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type NewsProvider interface {
	GetNews(ctx context.Context, ticker string, limit int) (domain.TickerNewsReport, error)
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
) (domain.TickerNewsReport, error) {
	headlines, err := s.provider.GetNews(ctx, ticker, newsLimit)
	if err != nil {
		return domain.TickerNewsReport{}, fmt.Errorf("get news for %s: %w", ticker, err)
	}

	return headlines, nil
}
