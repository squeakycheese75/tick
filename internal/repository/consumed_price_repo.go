package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
)

type ConsumedPriceRepository struct {
	db *sql.DB
	q  *db.Queries
}

func NewConsumedPriceRepository(database *db.DB) *ConsumedPriceRepository {
	return &ConsumedPriceRepository{
		q:  db.New(database.SqlDB),
		db: database.SqlDB}
}

func (r *ConsumedPriceRepository) Create(
	ctx context.Context,
	price ConsumedPrice,
) error {
	return r.q.UpsertConsumedPrice(ctx, db.UpsertConsumedPriceParams{
		Symbol:   strings.ToUpper(strings.TrimSpace(price.Symbol)),
		Source:   strings.TrimSpace(price.Source),
		Price:    price.Price,
		Currency: strings.ToUpper(strings.TrimSpace(price.Currency)),
		AsOf:     price.AsOf,
	})
}

func (r *ConsumedPriceRepository) GetLatest(
	ctx context.Context,
	symbol string,
) (ConsumedPrice, error) {
	row, err := r.q.GetLatestConsumerPrice(
		ctx,
		strings.ToUpper(strings.TrimSpace(symbol)),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConsumedPrice{}, domain.ErrConsumedPriceNotFound
		}

		return ConsumedPrice{}, err
	}

	return ConsumedPrice{
		Symbol:   row.Symbol,
		Source:   row.Source,
		Price:    row.Price,
		Currency: row.Currency,
		AsOf:     row.AsOf,
	}, nil
}
