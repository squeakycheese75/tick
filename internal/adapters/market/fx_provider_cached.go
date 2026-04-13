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

type FXProvider interface {
	GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (domain.FXRate, error)
}

type FXCacheStore interface {
	Get(ctx context.Context, baseCurrency, quoteCurrency string) (repository.CachedFXRate, error)
	Upsert(ctx context.Context, cached repository.CachedFXRate) error
}

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

	case !errors.Is(err, domain.ErrFXCacheNotFound):
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
