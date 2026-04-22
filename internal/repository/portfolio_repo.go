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

type PortfolioRepository struct {
	q *db.Queries
}

func NewPortfolioRepository(database *db.DB) *PortfolioRepository {
	return &PortfolioRepository{q: db.New(database.SqlDB)}
}

func (r *PortfolioRepository) GetByName(ctx context.Context, name string) (Portfolio, error) {
	row, err := r.q.GetPortfolioByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Portfolio{}, fmt.Errorf("portfolio %q: %w", name, domain.ErrPortfolioNotFound)
		}

		return Portfolio{}, fmt.Errorf("get portfolio by name %q: %w", name, err)
	}

	return Portfolio{
		ID:           row.ID,
		Name:         row.Name,
		BaseCurrency: row.BaseCurrency,
	}, nil
}

func (r *PortfolioRepository) Create(ctx context.Context, p Portfolio) error {
	err := r.q.CreatePortfolio(ctx, db.CreatePortfolioParams{
		Name:         p.Name,
		BaseCurrency: p.BaseCurrency,
	})
	if err != nil {
		return fmt.Errorf("create portfolio: %w", err)
	}

	return nil
}

func (r *PortfolioSnapshotRepository) GetLatestBefore(
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

func (r *PortfolioSnapshotRepository) ListPositionsBySnapshotID(
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
