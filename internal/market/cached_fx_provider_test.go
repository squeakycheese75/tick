package market_test

import (
	"context"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/market"
	"github.com/squeakycheese75/tick/internal/repository"
)

type fakeFXStore struct {
	rows map[string]repository.CachedFXRate
}

func (s *fakeFXStore) Get(_ context.Context, base, quote string) (repository.CachedFXRate, error) {
	row, ok := s.rows[base+quote]
	if !ok {
		return repository.CachedFXRate{}, domain.ErrFXCacheNotFound
	}
	return row, nil
}

func (s *fakeFXStore) Upsert(_ context.Context, c repository.CachedFXRate) error {
	s.rows[c.BaseCurrency+c.QuoteCurrency] = c
	return nil
}

type fakeFXProvider struct{ calls int }

func (p *fakeFXProvider) GetRate(_ context.Context, base, quote string) (domain.FXRate, error) {
	p.calls++
	return domain.FXRate{BaseCurrency: base, QuoteCurrency: quote, Rate: 1.12, Source: "fake"}, nil
}

func TestCachedFXProvider_GetRate_FetchesAndCachesOnMiss(t *testing.T) {
	t.Parallel()

	store := &fakeFXStore{rows: map[string]repository.CachedFXRate{}}
	inner := &fakeFXProvider{}
	provider := market.NewCachedFXProvider(inner, store, time.Hour)

	for i := 0; i < 2; i++ {
		rate, err := provider.GetRate(context.Background(), "eur", "usd")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if rate.Rate != 1.12 {
			t.Fatalf("call %d: rate = %v, want 1.12", i, rate.Rate)
		}
	}

	if inner.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1 (second call should hit the cache)", inner.calls)
	}
}

type fakePriceStore struct {
	rows map[string]repository.CachedPriceQuote
}

func (s *fakePriceStore) Get(_ context.Context, key string) (repository.CachedPriceQuote, error) {
	row, ok := s.rows[key]
	if !ok {
		return repository.CachedPriceQuote{}, domain.ErrPriceCacheNotFound
	}
	return row, nil
}

func (s *fakePriceStore) Upsert(_ context.Context, q repository.PriceQuote, at time.Time) error {
	s.rows[q.Symbol] = repository.CachedPriceQuote{PriceQuote: q, FetchedAt: at}
	return nil
}

type fakePriceProvider struct {
	asOf  time.Time
	calls int
}

func (p *fakePriceProvider) GetQuote(_ context.Context, in market.GetQuoteParams) (domain.Quote, error) {
	p.calls++
	return domain.Quote{Symbol: in.Symbol, Price: 10, PriceCurrency: "USD", AsOf: p.asOf}, nil
}

func TestCachedPriceProvider_GetQuote_KeepsTheTradeTimeAcrossTheCache(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	inner := &fakePriceProvider{asOf: asOf}
	provider := market.NewCachedPriceProvider(inner, &fakePriceStore{rows: map[string]repository.CachedPriceQuote{}}, time.Hour)

	for i := 0; i < 2; i++ {
		q, err := provider.GetQuote(context.Background(), market.GetQuoteParams{Symbol: "MRNA", ProviderSymbol: "MRNA"})
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if !q.AsOf.Equal(asOf) {
			t.Fatalf("call %d: AsOf = %v, want %v", i, q.AsOf, asOf)
		}
	}

	if inner.calls != 1 {
		t.Fatalf("upstream calls = %d, want 1 (second call should come from the cache)", inner.calls)
	}
}
