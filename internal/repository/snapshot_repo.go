package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
)

type SnapshotRepository struct {
	db *sql.DB
	q  *db.Queries
}

func NewSnapshotRepository(database *db.DB) *SnapshotRepository {
	return &SnapshotRepository{
		q:  db.New(database.SqlDB),
		db: database.SqlDB}
}

func (r *SnapshotRepository) Create(
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

func (r *SnapshotRepository) GetLatestBefore(
	ctx context.Context,
	portfolioName string,
	before time.Time,
) (PortfolioSnapshot, error) {
	row, err := r.q.GetLatestPortfolioSnapshotBefore(ctx, db.GetLatestPortfolioSnapshotBeforeParams{
		PortfolioName: portfolioName,
		CapturedAt:    before,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PortfolioSnapshot{}, domain.ErrPortfolioSnapshotNotFound
		}
		return PortfolioSnapshot{}, fmt.Errorf("get latest portfolio snapshot before: %w", err)
	}

	return PortfolioSnapshot{
		ID:            row.ID,
		PortfolioName: row.PortfolioName,
		BaseCurrency:  row.BaseCurrency,
		TotalValue:    row.TotalValue,
		CapturedAt:    row.CapturedAt,
	}, nil
}

func (r *SnapshotRepository) ListPositionsBySnapshotID(
	ctx context.Context,
	snapshotID int64,
) ([]PortfolioSnapshotPosition, error) {

	rows, err := r.q.ListPortfolioSnapshotPositionsBySnapshotID(ctx, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("list snapshot positions: %w", err)
	}

	out := make([]PortfolioSnapshotPosition, 0, len(rows))

	for _, row := range rows {
		out = append(out, PortfolioSnapshotPosition{
			SnapshotID:         row.SnapshotID,
			Symbol:             row.Symbol,
			Quantity:           row.Quantity,
			InstrumentCurrency: row.InstrumentCurrency,
			QuotedPrice:        row.QuotedPrice,
			FXRate:             row.FxRate,
			MarketValueBase:    row.MarketValueBase,
			Weight:             row.Weight,
		})
	}

	return out, nil
}
