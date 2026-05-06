package market

import (
	"context"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type (
	PriceCacheStore interface {
		Upsert(ctx context.Context, quote repository.PriceQuote, fetchedAt time.Time) error
		Get(ctx context.Context, symbol string) (repository.CachedPriceQuote, error)
	}
	PriceProvider interface {
		GetQuote(ctx context.Context, p GetQuoteParams) (domain.Quote, error)
	}
	FXCacheStore interface {
		Get(ctx context.Context, baseCurrency, quoteCurrency string) (repository.CachedFXRate, error)
		Upsert(ctx context.Context, cached repository.CachedFXRate) error
	}
	FXProvider interface {
		GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (domain.FXRate, error)
	}
	SymbolResolver interface {
		Resolve(symbol, provider string) (string, error)
	}
)

type GetQuoteParams struct {
	Symbol         string
	ProviderSymbol string
}
