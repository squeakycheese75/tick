package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
)

type TargetRepository struct {
	db *sql.DB
	q  *db.Queries
}

func NewTargetRepository(database *db.DB) *TargetRepository {
	return &TargetRepository{
		q:  db.New(database.SqlDB),
		db: database.SqlDB}
}

func (r *TargetRepository) Save(ctx context.Context, t domain.Target) error {
	_, err := r.q.CreateTarget(ctx, db.CreateTargetParams{
		PortfolioID:   t.PortfolioID,
		Symbol:        t.Symbol,
		Type:          string(t.Type),
		TargetPrice:   t.TargetPrice,
		QuoteCurrency: t.QuoteCurrency,
	})
	if err != nil {
		return fmt.Errorf(
			"create target for portfolio id %d and symbol %v: %w",
			t.PortfolioID,
			t.Symbol,
			err,
		)
	}

	return nil
}

func (r *TargetRepository) ListByPortfolio(ctx context.Context, portfolioID int64) ([]domain.Target, error) {
	rows, err := r.q.ListTargetsByPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("list positions by portfolio id %d: %w", portfolioID, err)
	}

	targets := make([]domain.Target, 0, len(rows))
	for _, row := range rows {
		targets = append(targets, domain.Target{
			Symbol:        row.Symbol,
			Type:          domain.TargetType(row.Type),
			QuoteCurrency: row.QuoteCurrency,
			TargetPrice:   row.TargetPrice,
		})
	}

	return targets, nil
}
