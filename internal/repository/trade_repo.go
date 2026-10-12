package repository

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
)

type TradeRepository struct {
	q *db.Queries
}

func NewTradeRepository(database *db.DB) *TradeRepository {
	return &TradeRepository{q: db.New(database.SqlDB)}
}

// Exists reports whether a trade with the given key was already recorded for the portfolio.
func (r *TradeRepository) Exists(ctx context.Context, portfolioID int64, key string) (bool, error) {
	n, err := r.q.CountTradesByKey(ctx, db.CountTradesByKeyParams{
		PortfolioID: portfolioID,
		TradeKey:    key,
	})
	if err != nil {
		return false, fmt.Errorf("check trade %q in portfolio id %d: %w", key, portfolioID, err)
	}

	return n > 0, nil
}

// Create records a trade in the ledger.
func (r *TradeRepository) Create(ctx context.Context, t Trade) error {
	err := r.q.CreateTrade(ctx, db.CreateTradeParams{
		PortfolioID:  t.PortfolioID,
		InstrumentID: t.InstrumentID,
		TradeKey:     t.Key,
		Side:         t.Side,
		Quantity:     t.Quantity,
		Price:        t.Price,
		Currency:     t.Currency,
		Fee:          t.Fee,
		FxRate:       t.FXRate,
		TradedAt:     t.TradedAt,
	})
	if err != nil {
		return fmt.Errorf("create trade %q: %w", t.Key, err)
	}

	return nil
}
