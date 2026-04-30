package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
)

type InstrumentRepository struct {
	q *db.Queries
}

func NewInstrumentRepository(database *db.DB) *InstrumentRepository {
	return &InstrumentRepository{q: db.New(database.SqlDB)}
}

func (r *InstrumentRepository) GetBySymbolAndExchange(
	ctx context.Context,
	symbol string,
	exchange string,
) (Instrument, error) {
	row, err := r.q.GetInstrumentBySymbolAndExchange(ctx, db.GetInstrumentBySymbolAndExchangeParams{
		Symbol: symbol,
		Exchange: sql.NullString{
			String: exchange,
			Valid:  exchange != "",
		},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Instrument{}, fmt.Errorf(
				"instrument %q on exchange %q: %w",
				symbol,
				exchange,
				domain.ErrInstrumentNotFound,
			)
		}

		return Instrument{}, fmt.Errorf(
			"get instrument by symbol %q and exchange %q: %w",
			symbol,
			exchange,
			err,
		)
	}

	return Instrument{
		ID:             row.ID,
		Symbol:         row.Symbol,
		ProviderSymbol: row.ProviderSymbol,
		Exchange:       row.Exchange.String,
		InstrumentType: row.AssetType,
		QuoteCurrency:  row.QuoteCurrency,
	}, nil
}

func (r *InstrumentRepository) Create(ctx context.Context, in Instrument) (Instrument, error) {
	id, err := r.q.CreateInstrument(ctx, db.CreateInstrumentParams{
		Symbol:         in.Symbol,
		ProviderSymbol: in.ProviderSymbol,
		AssetType:      in.InstrumentType,
		Exchange: sql.NullString{
			String: in.Exchange,
			Valid:  in.Exchange != "",
		},
		QuoteCurrency: in.QuoteCurrency,
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return Instrument{}, fmt.Errorf(
				"instrument %q on exchange %q: %w",
				in.Symbol,
				in.Exchange,
				domain.ErrPositionAlreadyExists,
			)
		}

		return Instrument{}, fmt.Errorf(
			"create instrument %q on exchange %q: %w",
			in.Symbol,
			in.Exchange,
			err,
		)
	}

	return Instrument{
		ID:             id,
		Symbol:         in.Symbol,
		ProviderSymbol: in.ProviderSymbol,
		InstrumentType: in.InstrumentType,
		Exchange:       in.Exchange,
		QuoteCurrency:  in.QuoteCurrency,
	}, nil
}

func (r *InstrumentRepository) GetOrCreate(ctx context.Context, in Instrument) (Instrument, error) {
	instrument, err := r.GetBySymbolAndExchange(ctx, in.Symbol, in.Exchange)
	if err == nil {
		return instrument, nil
	}

	if !errors.Is(err, domain.ErrInstrumentNotFound) {
		return Instrument{}, err
	}

	created, err := r.Create(ctx, in)
	if err == nil {
		return created, nil
	}

	if errors.Is(err, domain.ErrInstrumentAlreadyExists) {
		return r.GetBySymbolAndExchange(ctx, in.Symbol, in.Exchange)
	}

	return Instrument{}, err
}
