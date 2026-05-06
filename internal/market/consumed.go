package market

import (
	"context"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type (
	ConsumedPriceRepository interface {
		GetLatest(ctx context.Context, symbol string) (repository.ConsumedPrice, error)
	}
)

type ConsumedPriceProvider struct {
	repo   ConsumedPriceRepository
	maxAge time.Duration
}

func NewConsumedPriceProvider(repo ConsumedPriceRepository, maxAge time.Duration) *ConsumedPriceProvider {
	return &ConsumedPriceProvider{
		repo:   repo,
		maxAge: maxAge,
	}
}

func (p *ConsumedPriceProvider) GetQuote(
	ctx context.Context,
	in GetQuoteParams,
) (domain.Quote, error) {
	price, err := p.repo.GetLatest(ctx, in.Symbol)
	if err != nil {
		return domain.Quote{}, err
	}

	stale := time.Since(price.AsOf) > p.maxAge

	return domain.Quote{
		Symbol:        price.Symbol,
		Price:         price.Price,
		PriceCurrency: price.Currency,
		Source:        price.Source,
		AsOf:          price.AsOf,
		Stale:         stale,
	}, nil
}
