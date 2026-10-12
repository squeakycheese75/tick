package usecase_test

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/instruments"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase"
)

// These run the real use case on a real SQLite file. They open databases, and goose keeps its
// settings in package globals, so they cannot run in parallel with each other.

// sqliteUnitOfWork is the same translation the composition root does: the repository returns
// concrete types, and the use case wants its own interfaces.
type sqliteUnitOfWork struct{ uow *repository.TradeUnitOfWork }

func (s sqliteUnitOfWork) Reader() usecase.TradeStore { return s.uow.Reader() }

func (s sqliteUnitOfWork) Within(ctx context.Context, fn func(usecase.TradeStore) error) error {
	return s.uow.Within(ctx, func(store *repository.TradeStore) error { return fn(store) })
}

type fixedResolver struct{}

func (fixedResolver) Resolve(context.Context, string) (instruments.ResolvedInstrument, error) {
	return instruments.ResolvedInstrument{InstrumentType: "equity", Exchange: "NASDAQ", QuoteCurrency: "USD"}, nil
}

type fixedRates struct{}

func (fixedRates) GetRate(context.Context, string, string) (domain.FXRate, error) {
	return domain.FXRate{Rate: 0.5}, nil
}

type sqliteFixture struct {
	database  *db.DB
	uow       sqliteUnitOfWork
	portfolio repository.Portfolio
}

func newSQLiteFixture(t *testing.T) sqliteFixture {
	t.Helper()

	database, err := db.OpenAndMigrateSqlite(filepath.Join(t.TempDir(), "tick.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	portfolios := repository.NewPortfolioRepository(database)
	if err := portfolios.Create(context.Background(), repository.Portfolio{Name: "trending", BaseCurrency: "EUR"}); err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	portfolio, err := portfolios.GetByName(context.Background(), "trending")
	if err != nil {
		t.Fatalf("get portfolio: %v", err)
	}

	return sqliteFixture{database: database, uow: sqliteUnitOfWork{repository.NewTradeUnitOfWork(database)}, portfolio: portfolio}
}

func (f sqliteFixture) importer() *usecase.ImportTradesUseCase {
	return usecase.NewImportTradesUseCase(f.uow, fixedResolver{}, fixedRates{})
}

func (f sqliteFixture) position(t *testing.T) repository.Position {
	t.Helper()

	instrument, err := repository.NewInstrumentRepository(f.database).GetBySymbolAndExchange(context.Background(), "AAPL", "NASDAQ")
	if err != nil {
		t.Fatalf("get instrument: %v", err)
	}
	pos, err := repository.NewPositionRepository(f.database).Get(context.Background(), f.portfolio.ID, instrument.ID)
	if err != nil {
		t.Fatalf("get position: %v", err)
	}

	return pos
}

func (f sqliteFixture) ledgerRows(t *testing.T) int {
	t.Helper()

	var n int
	if err := f.database.SqlDB.QueryRow("SELECT COUNT(*) FROM trades").Scan(&n); err != nil {
		t.Fatalf("count trades: %v", err)
	}

	return n
}

func TestImportTrades_SQLite_ConcurrentImportsDoNotLoseUpdates(t *testing.T) {
	f := newSQLiteFixture(t)

	// Ten importers, each buying one more share on its own day, all at once. If they planned
	// from the same starting position and wrote over each other, shares would go missing.
	const importers = 10

	var wg sync.WaitGroup
	errs := make(chan error, importers)

	for i := 1; i <= importers; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			tr := domain.TradeInput{
				Date: time.Date(2026, 10, i, 0, 0, 0, 0, time.UTC), Symbol: "AAPL", Side: domain.TradeSideBuy,
				Quantity: 1, Price: float64(i * 10), Currency: "USD",
			}
			_, err := f.importer().Execute(context.Background(), domain.ImportTradesInput{PortfolioName: "trending", Trades: []domain.TradeInput{tr}})
			errs <- err
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("an importer failed: %v", err)
		}
	}

	// Prices 10..100 average to 55; the stored base cost is that at the fixed 0.5 rate.
	pos := f.position(t)
	if pos.Quantity != importers || math.Abs(pos.AvgCost-55) > 1e-9 || math.Abs(pos.AvgCostBase-27.5) > 1e-9 {
		t.Fatalf("position = %+v, want %d @ 55 (base 27.5): updates were lost", pos, importers)
	}
	if n := f.ledgerRows(t); n != importers {
		t.Fatalf("ledger rows = %d, want %d", n, importers)
	}
}

func TestImportTrades_SQLite_TheSameFileImportedTwiceAtOnceAppliesOnce(t *testing.T) {
	f := newSQLiteFixture(t)

	file := domain.ImportTradesInput{PortfolioName: "trending"}
	for i := 1; i <= 5; i++ {
		file.Trades = append(file.Trades, domain.TradeInput{
			Date: time.Date(2026, 10, i, 0, 0, 0, 0, time.UTC), Symbol: "AAPL", Side: domain.TradeSideBuy,
			Quantity: 2, Price: 100, Currency: "USD",
		})
	}

	// Each importer gets its own copy: Execute normalises the trades it is given.
	copyOf := func() domain.ImportTradesInput {
		in := file
		in.Trades = append([]domain.TradeInput(nil), file.Trades...)
		return in
	}

	var wg sync.WaitGroup
	outs := make([]*domain.ImportTradesOutput, 2)
	errs := make([]error, 2)

	for i := range outs {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			outs[i], errs[i] = f.importer().Execute(context.Background(), copyOf())
		}(i)
	}

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			t.Fatalf("an importer failed: %v", err)
		}
	}

	if applied := outs[0].Applied + outs[1].Applied; applied != 5 {
		t.Fatalf("applied %d + %d trades, want 5 in total (each trade exactly once)", outs[0].Applied, outs[1].Applied)
	}
	if skipped := outs[0].Skipped + outs[1].Skipped; skipped != 5 {
		t.Fatalf("skipped = %d, want 5", skipped)
	}

	if pos := f.position(t); pos.Quantity != 10 {
		t.Fatalf("quantity = %v, want 10 (not doubled)", pos.Quantity)
	}
	if n := f.ledgerRows(t); n != 5 {
		t.Fatalf("ledger rows = %d, want 5", n)
	}
}

func TestTradeUnitOfWork_SQLite_FailureRollsEverythingBack(t *testing.T) {
	f := newSQLiteFixture(t)
	ctx := context.Background()

	err := f.uow.uow.Within(ctx, func(store *repository.TradeStore) error {
		inst, err := store.GetOrCreateInstrument(ctx, repository.Instrument{Symbol: "AAPL", InstrumentType: "equity", Exchange: "NASDAQ", QuoteCurrency: "USD"})
		if err != nil {
			return err
		}
		if err := store.SavePosition(ctx, repository.CreatePositionParams{InstrumentID: inst.ID, PortfolioID: f.portfolio.ID, Quantity: 1, AvgCost: 1, Currency: "USD"}); err != nil {
			return err
		}

		return fmt.Errorf("fail after writing")
	})
	if err == nil {
		t.Fatal("expected the error to come back")
	}

	if _, err := repository.NewInstrumentRepository(f.database).GetBySymbolAndExchange(ctx, "AAPL", "NASDAQ"); err == nil {
		t.Fatal("the instrument written inside the failed transaction survived")
	}

	var positions int
	if err := f.database.SqlDB.QueryRow("SELECT COUNT(*) FROM positions").Scan(&positions); err != nil || positions != 0 {
		t.Fatalf("positions = %d err %v, want 0", positions, err)
	}
}

func TestTradeUnitOfWork_SQLite_PanicRollsBackAndPropagates(t *testing.T) {
	f := newSQLiteFixture(t)
	ctx := context.Background()

	func() {
		defer func() {
			if recover() == nil {
				t.Error("the panic should propagate")
			}
		}()

		_ = f.uow.uow.Within(ctx, func(store *repository.TradeStore) error {
			if _, err := store.GetOrCreateInstrument(ctx, repository.Instrument{Symbol: "AAPL", InstrumentType: "equity", Exchange: "NASDAQ", QuoteCurrency: "USD"}); err != nil {
				return err
			}
			panic("boom")
		})
	}()

	if _, err := repository.NewInstrumentRepository(f.database).GetBySymbolAndExchange(ctx, "AAPL", "NASDAQ"); err == nil {
		t.Fatal("a panic must roll back what was written")
	}

	// The lock must have been released: a later write has to go through, not wait forever.
	done := make(chan error, 1)
	go func() {
		done <- f.uow.uow.Within(ctx, func(*repository.TradeStore) error { return nil })
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("a later transaction failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("a later transaction is stuck: the failed one kept the write lock")
	}
}
