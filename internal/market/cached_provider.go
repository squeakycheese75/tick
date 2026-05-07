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

type CachedFXProvider struct {
	inner FXProvider
	repo  FXCacheStore
	ttl   time.Duration
}

func NewCachedFXProvider(
	inner FXProvider,
	repo FXCacheStore,
	ttl time.Duration,
) *CachedFXProvider {
	return &CachedFXProvider{
		inner: inner,
		repo:  repo,
		ttl:   ttl,
	}
}

func (p *CachedFXProvider) GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (domain.FXRate, error) {
	base := strings.ToUpper(strings.TrimSpace(baseCurrency))
	quote := strings.ToUpper(strings.TrimSpace(quoteCurrency))
	now := time.Now()

	cached, err := p.repo.Get(ctx, base, quote)
	switch {
	case err == nil:
		if now.Sub(cached.FetchedAt) < p.ttl {
			return toDomainFXRate(cached), nil
		}

	case !errors.Is(err, domain.ErrFXRateNotFound):
		return domain.FXRate{}, fmt.Errorf("get cached fx rate for %s/%s: %w", base, quote, err)
	}

	rate, err := p.inner.GetRate(ctx, base, quote)
	if err != nil {
		return domain.FXRate{}, err
	}

	_ = p.repo.Upsert(ctx, toRepositoryFXRate(rate, now))

	return rate, nil
}

func toDomainFXRate(c repository.CachedFXRate) domain.FXRate {
	return domain.FXRate{
		BaseCurrency:  c.BaseCurrency,
		QuoteCurrency: c.QuoteCurrency,
		Rate:          c.Rate,
		Source:        c.Source,
	}
}

func toRepositoryFXRate(r domain.FXRate, fetchedAt time.Time) repository.CachedFXRate {
	return repository.CachedFXRate{
		BaseCurrency:  r.BaseCurrency,
		QuoteCurrency: r.QuoteCurrency,
		Rate:          r.Rate,
		Source:        r.Source,
		FetchedAt:     fetchedAt,
	}
}

type CachedPriceProvider struct {
	inner PriceProvider
	ttl   time.Duration
	repo  PriceCacheStore
}

func NewCachedPriceProvider(inner PriceProvider, cacheRepo PriceCacheStore, ttl time.Duration) *CachedPriceProvider {
	return &CachedPriceProvider{
		inner: inner,
		ttl:   ttl,
		repo:  cacheRepo,
	}
}

func (p *CachedPriceProvider) GetQuote(ctx context.Context, in GetQuoteParams) (domain.Quote, error) {
	key := strings.ToUpper(strings.TrimSpace(in.Symbol))
	keyProviderSymbol := strings.ToUpper(strings.TrimSpace(in.ProviderSymbol))
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
	quote, err := p.inner.GetQuote(ctx, GetQuoteParams{
		Symbol:         key,
		ProviderSymbol: keyProviderSymbol,
	})
	if err != nil {
		return domain.Quote{}, err
	}

	// 3. Store in cache (best effort)
	err = p.repo.Upsert(ctx, toRepositoryQuote(quote, in), now)
	if err != nil {
		return domain.Quote{}, err
	}

	return quote, nil
}

func toDomainQuote(c repository.CachedPriceQuote) domain.Quote {
	return domain.Quote{
		Symbol:        c.PriceQuote.Symbol,
		Price:         c.PriceQuote.Price,
		PriceCurrency: c.PriceQuote.PriceCurrency,
		PreviousClose: c.PriceQuote.PreviousClose,
		Change:        c.PriceQuote.Change,
		ChangePercent: c.PriceQuote.ChangePercent,
		Source:        c.PriceQuote.Source,
	}
}

func toRepositoryQuote(q domain.Quote, in GetQuoteParams) repository.PriceQuote {
	return repository.PriceQuote{
		Symbol:        in.Symbol,
		SourceSymbol:  in.ProviderSymbol,
		Price:         q.Price,
		PriceCurrency: q.PriceCurrency,
		PreviousClose: q.PreviousClose,
		Change:        q.Change,
		ChangePercent: q.ChangePercent,
		Source:        q.Source,
	}
}
