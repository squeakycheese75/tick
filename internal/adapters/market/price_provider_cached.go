package market

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

type CachedPriceProvider struct {
	inner PriceProvider
	ttl   time.Duration

	mu    sync.RWMutex
	cache map[string]cachedQuote
}

type cachedQuote struct {
	quote     domain.Quote
	expiresAt time.Time
}

type PriceProvider interface {
	GetQuote(ctx context.Context, ticker string) (domain.Quote, error)
}

func NewCachedPriceProvider(inner PriceProvider, ttl time.Duration) *CachedPriceProvider {
	return &CachedPriceProvider{
		inner: inner,
		ttl:   ttl,
		cache: make(map[string]cachedQuote),
	}
}

func (p *CachedPriceProvider) GetQuote(ctx context.Context, ticker string) (domain.Quote, error) {
	key := strings.ToUpper(strings.TrimSpace(ticker))
	now := time.Now()

	p.mu.RLock()
	entry, ok := p.cache[key]
	p.mu.RUnlock()

	if ok && now.Before(entry.expiresAt) {
		return entry.quote, nil
	}

	quote, err := p.inner.GetQuote(ctx, key)
	if err != nil {
		return domain.Quote{}, err
	}

	p.mu.Lock()
	p.cache[key] = cachedQuote{
		quote:     quote,
		expiresAt: now.Add(p.ttl),
	}
	p.mu.Unlock()

	return quote, nil
}
