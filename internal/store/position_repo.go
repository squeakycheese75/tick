package store

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type PositionRepository struct {
	db *DB
}

func NewPositionRepository(db *DB) *PositionRepository {
	return &PositionRepository{db: db}
}

func (r *PositionRepository) ListByPortfolio(ctx context.Context, portfolioName string) ([]domain.Position, error) {
	const query = `
SELECT portfolio_name, ticker, quantity, avg_cost, currency
FROM positions
WHERE portfolio_name = ?
ORDER BY ticker ASC;
`

	rows, err := r.db.sqlDB.QueryContext(ctx, query, portfolioName)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer rows.Close()

	positions := make([]domain.Position, 0)
	for rows.Next() {
		var p domain.Position
		if err := rows.Scan(&p.PortfolioName, &p.Ticker, &p.Quantity, &p.AvgCost, &p.InstrumentCurrency); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		positions = append(positions, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions: %w", err)
	}

	return positions, nil
}

func (r *PositionRepository) Upsert(ctx context.Context, p domain.Position) error {
	const query = `
INSERT INTO positions (portfolio_name, ticker, quantity, avg_cost, currency)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(portfolio_name, ticker) DO UPDATE SET
    quantity = excluded.quantity,
    avg_cost = excluded.avg_cost,
    currency = excluded.currency;
`

	_, err := r.db.sqlDB.ExecContext(
		ctx,
		query,
		p.PortfolioName,
		p.Ticker,
		p.Quantity,
		p.AvgCost,
		p.InstrumentCurrency,
	)
	if err != nil {
		return fmt.Errorf("upsert position: %w", err)
	}

	return nil
}
