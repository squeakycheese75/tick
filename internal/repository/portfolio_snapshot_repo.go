package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
)

type PortfolioSnapshotRepository struct {
	db *sql.DB
	q  *db.Queries
}

func NewPortfolioSnapshotRepository(database *db.DB) *PortfolioSnapshotRepository {
	return &PortfolioSnapshotRepository{
		q:  db.New(database.SqlDB),
		db: database.SqlDB}
}

func (r *PortfolioSnapshotRepository) Create(
	ctx context.Context,
	in PortfolioSnapshot,
	positions []PortfolioSnapshotPosition,
) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	q := r.q.WithTx(tx)

	row, err := q.CreatePortfolioSnapshot(ctx, db.CreatePortfolioSnapshotParams{
		PortfolioName: in.PortfolioName,
		BaseCurrency:  in.BaseCurrency,
		TotalValue:    in.TotalValue,
		CapturedAt:    in.CapturedAt,
	})
	if err != nil {
		return 0, fmt.Errorf("create portfolio snapshot: %w", err)
	}

	for _, p := range positions {
		_, err := q.CreatePortfolioSnapshotPosition(ctx, db.CreatePortfolioSnapshotPositionParams{
			SnapshotID:         row.ID,
			Symbol:             p.Symbol,
			Quantity:           p.Quantity,
			InstrumentCurrency: p.InstrumentCurrency,
			QuotedPrice:        p.QuotedPrice,
			FxRate:             p.FXRate,
			MarketValueBase:    p.MarketValueBase,
			Weight:             p.Weight,
		})
		if err != nil {
			return 0, fmt.Errorf("create portfolio snapshot position for %s: %w", p.Symbol, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return row.ID, nil
}
