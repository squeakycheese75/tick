package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type PortfolioRepository struct {
	db *DB
}

func NewPortfolioRepository(db *DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

func (r *PortfolioRepository) GetByName(ctx context.Context, name string) (domain.Portfolio, error) {
	const query = `
SELECT name, base_currency
FROM portfolios
WHERE name = ?;
`

	var p domain.Portfolio

	err := r.db.sqlDB.QueryRowContext(ctx, query, name).
		Scan(&p.Name, &p.BaseCurrency)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Portfolio{}, fmt.Errorf("portfolio %q not found", name)
		}
		return domain.Portfolio{}, fmt.Errorf("query portfolio: %w", err)
	}

	return p, nil
}

func (r *PortfolioRepository) Upsert(ctx context.Context, p domain.Portfolio) error {
	const query = `
INSERT INTO portfolios (name, base_currency)
VALUES (?, ?)
ON CONFLICT(name) DO UPDATE SET
    base_currency = excluded.base_currency;
`

	_, err := r.db.sqlDB.ExecContext(ctx, query, p.Name, p.BaseCurrency)
	if err != nil {
		return fmt.Errorf("upsert portfolio: %w", err)
	}

	return nil
}
