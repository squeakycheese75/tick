package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
)

type FXCacheRepository struct {
	q *db.Queries
}

func NewFXCacheRepository(database *db.DB) *FXCacheRepository {
	return &FXCacheRepository{q: db.New(database.SqlDB)}
}

func (r *FXCacheRepository) Get(ctx context.Context, baseCurrency, quoteCurrency string) (CachedFXRate, error) {
	row, err := r.q.GetFXCacheByPair(ctx, db.GetFXCacheByPairParams{
		BaseCurrency:  strings.ToUpper(strings.TrimSpace(baseCurrency)),
		QuoteCurrency: strings.ToUpper(strings.TrimSpace(quoteCurrency)),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CachedFXRate{}, domain.ErrFXCacheNotFound
		}
		return CachedFXRate{}, fmt.Errorf("get fx cache for %s/%s: %w", baseCurrency, quoteCurrency, err)
	}

	return CachedFXRate{
		BaseCurrency:  row.BaseCurrency,
		QuoteCurrency: row.QuoteCurrency,
		Rate:          row.Rate,
		Source:        row.Source,
		FetchedAt:     row.FetchedAt,
	}, nil
}

func (r *FXCacheRepository) Upsert(ctx context.Context, cached CachedFXRate) error {
	err := r.q.UpsertFXCache(ctx, db.UpsertFXCacheParams{
		BaseCurrency:  strings.ToUpper(strings.TrimSpace(cached.BaseCurrency)),
		QuoteCurrency: strings.ToUpper(strings.TrimSpace(cached.QuoteCurrency)),
		Rate:          cached.Rate,
		Source:        cached.Source,
		FetchedAt:     cached.FetchedAt,
	})
	if err != nil {
		return fmt.Errorf("upsert fx cache for %s/%s: %w", cached.BaseCurrency, cached.QuoteCurrency, err)
	}

	return nil
}
