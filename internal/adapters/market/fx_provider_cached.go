package market

import (
	"context"
	"strings"
	"sync"
	"time"
)

type FXProvider interface {
	GetRate(ctx context.Context, from string, to string) (float64, error)
}

type CachedFXProvider struct {
	inner FXProvider
	ttl   time.Duration

	mu    sync.RWMutex
	cache map[string]cachedRate
}

type cachedRate struct {
	rate      float64
	expiresAt time.Time
}

func NewCachedFXProvider(inner FXProvider, ttl time.Duration) *CachedFXProvider {
	return &CachedFXProvider{
		inner: inner,
		ttl:   ttl,
		cache: make(map[string]cachedRate),
	}
}

func (p *CachedFXProvider) GetRate(ctx context.Context, from string, to string) (float64, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))

	key := from + ":" + to
	now := time.Now()

	// fast path
	p.mu.RLock()
	entry, ok := p.cache[key]
	p.mu.RUnlock()

	if ok && now.Before(entry.expiresAt) {
		return entry.rate, nil
	}

	// fetch from inner provider
	rate, err := p.inner.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}

	// store
	p.mu.Lock()
	p.cache[key] = cachedRate{
		rate:      rate,
		expiresAt: now.Add(p.ttl),
	}
	p.mu.Unlock()

	return rate, nil
}
