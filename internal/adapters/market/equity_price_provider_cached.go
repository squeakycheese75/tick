package market

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type CachedPriceProvider struct {
	inner PriceProvider
	ttl   time.Duration
	repo  PriceCacheStore
}

type (
	PriceCacheStore interface {
		Upsert(ctx context.Context, quote repository.PriceQuote, fetchedAt time.Time) error
		Get(ctx context.Context, ticker string) (repository.CachedPriceQuote, error)
	}
	PriceProvider interface {
		GetQuote(ctx context.Context, ticker string) (domain.Quote, error)
	}
)

func NewCachedPriceProvider(inner PriceProvider, cacheRepo PriceCacheStore, ttl time.Duration) *CachedPriceProvider {
	return &CachedPriceProvider{
		inner: inner,
		ttl:   ttl,
		repo:  cacheRepo,
	}
}

func (p *CachedPriceProvider) GetQuote(ctx context.Context, ticker string) (domain.Quote, error) {
	key := strings.ToUpper(strings.TrimSpace(ticker))
	now := time.Now()

	// 1. Try cache
	cached, err := p.repo.Get(ctx, key)
	switch {
	case err == nil:
		if now.Sub(cached.FetchedAt) < p.ttl {
			return toDomainQuote(cached), nil
		}

	case !errors.Is(err, domain.ErrPriceCacheNotFound):
		return domain.Quote{}, fmt.Errorf("get cached quote for %q: %w", key, err)
	}

	// 2. Fetch fresh
	quote, err := p.inner.GetQuote(ctx, key)
	if err != nil {
		return domain.Quote{}, err
	}

	// 3. Store in cache (best effort)
	_ = p.repo.Upsert(ctx, toRepositoryQuote(quote), now)

	return quote, nil
}

func toDomainQuote(c repository.CachedPriceQuote) domain.Quote {
	return domain.Quote{
		Symbol:        c.PriceQuote.Ticker,
		Price:         c.PriceQuote.Price,
		PriceCurrency: c.PriceQuote.PriceCurrency,
		PreviousClose: c.PriceQuote.PreviousClose,
		Change:        c.PriceQuote.Change,
		ChangePercent: c.PriceQuote.ChangePercent,
		Source:        c.PriceQuote.Source,
	}
}

func toRepositoryQuote(q domain.Quote) repository.PriceQuote {
	return repository.PriceQuote{
		Ticker:        q.Symbol,
		Price:         q.Price,
		PriceCurrency: q.PriceCurrency,
		PreviousClose: q.PreviousClose,
		Change:        q.Change,
		ChangePercent: q.ChangePercent,
		Source:        q.Source,
	}
}
