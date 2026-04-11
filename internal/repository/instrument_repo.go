package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
)

type InstrumentRepository struct {
	q *db.Queries
}

func NewInstrumentRepository(database *db.DB) *InstrumentRepository {
	return &InstrumentRepository{q: db.New(database.SqlDB)}
}

func (r *InstrumentRepository) GetBySymbol(ctx context.Context, symbol string) (Instrument, error) {
	row, err := r.q.GetInstrumentBySymbol(ctx, symbol)
	if err != nil {
		return Instrument{}, fmt.Errorf("get instrument by symbol %v: %w", symbol, err)
	}

	return Instrument{
		ID:             row.ID,
		Symbol:         row.Symbol,
		ProviderSymbol: row.ProviderSymbol,
		Exchange:       row.Exchange.String,
		AssetType:      row.AssetType,
		QuoteCurrency:  row.QuoteCurrency,
	}, nil
}

func (r *InstrumentRepository) Create(ctx context.Context, p Instrument) error {
	err := r.q.CreateInstrument(ctx, db.CreateInstrumentParams{
		Symbol:         p.Symbol,
		ProviderSymbol: p.ProviderSymbol,
		AssetType:      p.AssetType,
		Exchange: sql.NullString{
			Valid:  true,
			String: p.Exchange,
		},
		QuoteCurrency: p.QuoteCurrency,
	})
	if err != nil {
		return fmt.Errorf("create portfolio: %w", err)
	}

	return nil
}
