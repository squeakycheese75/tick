package repository

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
)

type PositionRepository struct {
	q *db.Queries
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
			Quantity: row.Quantity,
			AvgCost:  row.AvgCost,
			Currency: row.Currency,
			Instrument: Instrument{
				Symbol:        row.Symbol,
				AssetType:     row.AssetType,
				QuoteCurrency: row.QuoteCurrency,
				Exchange:      row.Exchange.String,
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
