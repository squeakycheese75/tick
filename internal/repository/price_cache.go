package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
)

type PriceCacheRepository struct {
	q *db.Queries
}

func NewPriceCacheRepository(database *db.DB) *PriceCacheRepository {
	return &PriceCacheRepository{q: db.New(database.SqlDB)}
}

func (r *PriceCacheRepository) Get(ctx context.Context, ticker string) (CachedPriceQuote, error) {
	row, err := r.q.GetPriceCacheByTicker(ctx, strings.ToUpper(strings.TrimSpace(ticker)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CachedPriceQuote{}, domain.ErrPriceCacheNotFound
		}
		return CachedPriceQuote{}, fmt.Errorf("get price cache for %q: %w", ticker, err)
	}

	return CachedPriceQuote{
		PriceQuote: PriceQuote{
			Symbol:        row.Symbol,
			SourceSymbol:  row.ProviderSymbol.String,
			Price:         row.Price,
			PriceCurrency: row.PriceCurrency,
			PreviousClose: row.PreviousClose,
			Change:        row.Change,
			ChangePercent: row.ChangePercent,
			Source:        row.Source,
		},
		FetchedAt: row.FetchedAt,
	}, nil
}

func (r *PriceCacheRepository) Upsert(ctx context.Context, quote PriceQuote, fetchedAt time.Time) error {
	err := r.q.UpsertPriceCache(ctx, db.UpsertPriceCacheParams{
		Symbol: strings.ToUpper(strings.TrimSpace(quote.Symbol)),
		ProviderSymbol: sql.NullString{
			Valid:  true,
			String: strings.ToUpper(strings.TrimSpace(quote.SourceSymbol)),
		},
		Price:         quote.Price,
		PriceCurrency: quote.PriceCurrency,
		PreviousClose: quote.PreviousClose,
		Change:        quote.Change,
		ChangePercent: quote.ChangePercent,
		Source:        quote.Source,
		FetchedAt:     fetchedAt,
	})
	if err != nil {
		return fmt.Errorf("upsert price cache for %q: %w", quote.Symbol, err)
	}

	return nil
}
