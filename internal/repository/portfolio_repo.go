package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

// List returns every portfolio with its position count, ordered by name.
func (r *PortfolioRepository) List(ctx context.Context) ([]PortfolioOverview, error) {
	rows, err := r.q.ListPortfoliosWithPositionCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("list portfolios: %w", err)
	}

	portfolios := make([]PortfolioOverview, 0, len(rows))
	for _, row := range rows {
		portfolios = append(portfolios, PortfolioOverview{
			Name:          row.Name,
			BaseCurrency:  row.BaseCurrency,
			PositionCount: int(row.PositionCount),
		})
	}

	return portfolios, nil
}
