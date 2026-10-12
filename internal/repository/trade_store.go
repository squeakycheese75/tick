package repository

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/db"
)

// TradeStore is everything importing trades reads and writes, as one set of operations that
// either all run on the database or all run inside a single transaction.
type TradeStore struct {
	portfolios  *PortfolioRepository
	instruments *InstrumentRepository
	positions   *PositionRepository
	trades      *TradeRepository
}

// GetPortfolio returns the portfolio with the given name.
func (s *TradeStore) GetPortfolio(ctx context.Context, name string) (Portfolio, error) {
	return s.portfolios.GetByName(ctx, name)
}

// GetInstrument returns the instrument for a symbol on an exchange.
func (s *TradeStore) GetInstrument(ctx context.Context, symbol, exchange string) (Instrument, error) {
	return s.instruments.GetBySymbolAndExchange(ctx, symbol, exchange)
}

// GetOrCreateInstrument returns the instrument, creating it when it does not exist.
func (s *TradeStore) GetOrCreateInstrument(ctx context.Context, in Instrument) (Instrument, error) {
	return s.instruments.GetOrCreate(ctx, in)
}

// GetPosition returns what a portfolio holds of an instrument.
func (s *TradeStore) GetPosition(ctx context.Context, portfolioID, instrumentID int64) (Position, error) {
	return s.positions.Get(ctx, portfolioID, instrumentID)
}

// SavePosition creates or replaces a position.
func (s *TradeStore) SavePosition(ctx context.Context, p CreatePositionParams) error {
	return s.positions.Create(ctx, p)
}

// DeletePosition removes a position.
func (s *TradeStore) DeletePosition(ctx context.Context, portfolioID, instrumentID int64) error {
	return s.positions.Delete(ctx, portfolioID, instrumentID)
}

// TradeRecorded reports whether a trade with this key is already in the ledger.
func (s *TradeStore) TradeRecorded(ctx context.Context, portfolioID int64, key string) (bool, error) {
	return s.trades.Exists(ctx, portfolioID, key)
}

// RecordTrade appends a trade to the ledger.
func (s *TradeStore) RecordTrade(ctx context.Context, t Trade) error {
	return s.trades.Create(ctx, t)
}

// TradeUnitOfWork runs groups of trade operations atomically.
type TradeUnitOfWork struct {
	database *db.DB
	store    *TradeStore
}

// NewTradeUnitOfWork builds a TradeUnitOfWork on database.
func NewTradeUnitOfWork(database *db.DB) *TradeUnitOfWork {
	return &TradeUnitOfWork{
		database: database,
		store: &TradeStore{
			portfolios:  NewPortfolioRepository(database),
			instruments: NewInstrumentRepository(database),
			positions:   NewPositionRepository(database),
			trades:      NewTradeRepository(database),
		},
	}
}

// Reader returns a store that reads and writes directly, outside any transaction.
func (u *TradeUnitOfWork) Reader() *TradeStore {
	return u.store
}

// Within runs fn in one write transaction. The write lock is taken when it starts, so
// nothing else can change what fn reads, and everything fn writes is committed together or,
// if fn returns an error or panics, not at all.
func (u *TradeUnitOfWork) Within(ctx context.Context, fn func(*TradeStore) error) (err error) {
	tx, err := u.database.SqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	txStore := &TradeStore{
		portfolios:  u.store.portfolios.withTx(tx),
		instruments: u.store.instruments.withTx(tx),
		positions:   u.store.positions.withTx(tx),
		trades:      u.store.trades.withTx(tx),
	}

	if err := fn(txStore); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w (rollback also failed: %v)", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
