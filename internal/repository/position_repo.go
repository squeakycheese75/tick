package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
)

type PositionRepository struct {
	q *db.Queries
}

// withTx returns a repository that runs on tx.
func (r *PositionRepository) withTx(tx *sql.Tx) *PositionRepository {
	return &PositionRepository{q: r.q.WithTx(tx)}
}

func NewPositionRepository(database *db.DB) *PositionRepository {
	return &PositionRepository{q: db.New(database.SqlDB)}
}

func (r *PositionRepository) ListByPortfolioID(ctx context.Context, portfolioID int64) ([]Position, error) {
	rows, err := r.q.ListPositionsByPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("list positions by portfolio id %d: %w", portfolioID, err)
	}

	positions := make([]Position, 0, len(rows))
	for _, row := range rows {
		positions = append(positions, Position{
			Quantity:    row.Quantity,
			AvgCost:     row.AvgCost,
			Currency:    row.Currency,
			AvgCostBase: row.AvgCostBase,
			Instrument: Instrument{
				Symbol:         row.Symbol,
				InstrumentType: row.AssetType,
				QuoteCurrency:  row.QuoteCurrency,
				Exchange:       row.Exchange.String,
			},
		})
	}

	return positions, nil
}

func (r *PositionRepository) Create(
	ctx context.Context,
	p CreatePositionParams,
) error {
	err := r.q.CreatePosition(ctx, db.CreatePositionParams{
		PortfolioID:  p.PortfolioID,
		InstrumentID: p.InstrumentID,
		Quantity:     p.Quantity,
		AvgCost:      p.AvgCost,
		Currency:     p.Currency,
		AvgCostBase:  p.AvgCostBase,
	})
	if err != nil {
		return fmt.Errorf(
			"create position for portfolio id %d and instrument id %d: %w",
			p.PortfolioID,
			p.InstrumentID,
			err,
		)
	}

	return nil
}

// Get returns the position for an instrument in a portfolio, or domain.ErrPositionNotFound.
func (r *PositionRepository) Get(ctx context.Context, portfolioID, instrumentID int64) (Position, error) {
	row, err := r.q.GetPosition(ctx, db.GetPositionParams{
		PortfolioID:  portfolioID,
		InstrumentID: instrumentID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Position{}, fmt.Errorf(
			"position for portfolio id %d and instrument id %d: %w",
			portfolioID,
			instrumentID,
			domain.ErrPositionNotFound,
		)
	}
	if err != nil {
		return Position{}, fmt.Errorf(
			"get position for portfolio id %d and instrument id %d: %w",
			portfolioID,
			instrumentID,
			err,
		)
	}

	return Position{
		Quantity:    row.Quantity,
		AvgCost:     row.AvgCost,
		Currency:    row.Currency,
		AvgCostBase: row.AvgCostBase,
	}, nil
}

// Delete removes the position for an instrument in a portfolio.
func (r *PositionRepository) Delete(ctx context.Context, portfolioID, instrumentID int64) error {
	err := r.q.DeletePosition(ctx, db.DeletePositionParams{
		PortfolioID:  portfolioID,
		InstrumentID: instrumentID,
	})
	if err != nil {
		return fmt.Errorf(
			"delete position for portfolio id %d and instrument id %d: %w",
			portfolioID,
			instrumentID,
			err,
		)
	}

	return nil
}
