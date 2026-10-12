package repository_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

func TestTradesAndPositions_RoundTrip(t *testing.T) {
	// Not parallel: goose keeps its dialect and filesystem in package globals.
	ctx := context.Background()

	database, err := db.OpenAndMigrateSqlite(filepath.Join(t.TempDir(), "tick.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	portfolios := repository.NewPortfolioRepository(database)
	instruments := repository.NewInstrumentRepository(database)
	positions := repository.NewPositionRepository(database)
	trades := repository.NewTradeRepository(database)

	if err := portfolios.Create(ctx, repository.Portfolio{Name: "trending", BaseCurrency: "EUR"}); err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	portfolio, err := portfolios.GetByName(ctx, "trending")
	if err != nil {
		t.Fatalf("get portfolio: %v", err)
	}
	inst, err := instruments.GetOrCreate(ctx, repository.Instrument{
		Symbol: "MRNA", InstrumentType: "equity", Exchange: "NASDAQ", QuoteCurrency: "USD",
	})
	if err != nil {
		t.Fatalf("create instrument: %v", err)
	}

	t.Run("unknown position is not found", func(t *testing.T) {
		_, err := positions.Get(ctx, portfolio.ID, inst.ID)
		if !errors.Is(err, domain.ErrPositionNotFound) {
			t.Fatalf("err = %v, want ErrPositionNotFound", err)
		}
	})

	t.Run("position can be saved, read and deleted", func(t *testing.T) {
		err := positions.Create(ctx, repository.CreatePositionParams{
			InstrumentID: inst.ID, PortfolioID: portfolio.ID, Quantity: 2, AvgCost: 50, Currency: "USD", AvgCostBase: 45,
		})
		if err != nil {
			t.Fatalf("create position: %v", err)
		}

		got, err := positions.Get(ctx, portfolio.ID, inst.ID)
		if err != nil || got.Quantity != 2 || got.AvgCost != 50 || got.Currency != "USD" || got.AvgCostBase != 45 {
			t.Fatalf("got %+v err %v", got, err)
		}

		if err := positions.Delete(ctx, portfolio.ID, inst.ID); err != nil {
			t.Fatalf("delete position: %v", err)
		}
		if _, err := positions.Get(ctx, portfolio.ID, inst.ID); !errors.Is(err, domain.ErrPositionNotFound) {
			t.Fatalf("err = %v, want ErrPositionNotFound after delete", err)
		}
	})

	t.Run("trade keys are recorded and unique per portfolio", func(t *testing.T) {
		exists, err := trades.Exists(ctx, portfolio.ID, "k1")
		if err != nil || exists {
			t.Fatalf("exists = %v err %v, want false", exists, err)
		}

		trade := repository.Trade{
			PortfolioID: portfolio.ID, InstrumentID: inst.ID, Key: "k1", Side: "buy",
			Quantity: 1, Price: 10, Currency: "USD", FXRate: 1, TradedAt: time.Now(),
		}
		if err := trades.Create(ctx, trade); err != nil {
			t.Fatalf("create trade: %v", err)
		}

		exists, err = trades.Exists(ctx, portfolio.ID, "k1")
		if err != nil || !exists {
			t.Fatalf("exists = %v err %v, want true", exists, err)
		}

		if err := trades.Create(ctx, trade); err == nil {
			t.Fatal("recording the same key twice should fail")
		}
	})
}

func TestPriceCache_RoundTripsTheTradeTime(t *testing.T) {
	// Not parallel: goose keeps its dialect and filesystem in package globals.
	ctx := context.Background()

	database, err := db.OpenAndMigrateSqlite(filepath.Join(t.TempDir(), "tick.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	cache := repository.NewPriceCacheRepository(database)
	asOf := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)

	for name, in := range map[string]time.Time{"with a trade time": asOf, "without one": {}} {
		quote := repository.PriceQuote{Symbol: "MRNA", SourceSymbol: "MRNA", Price: 225, PriceCurrency: "USD", Source: "finnhub", AsOf: in}
		if err := cache.Upsert(ctx, quote, time.Now()); err != nil {
			t.Fatalf("%s: upsert: %v", name, err)
		}

		got, err := cache.Get(ctx, "MRNA")
		if err != nil {
			t.Fatalf("%s: get: %v", name, err)
		}
		if !got.PriceQuote.AsOf.Equal(in) {
			t.Fatalf("%s: AsOf = %v, want %v", name, got.PriceQuote.AsOf, in)
		}
	}
}

func TestPortfolioRepository_List_CountsPositions(t *testing.T) {
	// Not parallel: goose keeps its dialect and filesystem in package globals.
	ctx := context.Background()

	database, err := db.OpenAndMigrateSqlite(filepath.Join(t.TempDir(), "tick.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	portfolios := repository.NewPortfolioRepository(database)
	instruments := repository.NewInstrumentRepository(database)
	positions := repository.NewPositionRepository(database)

	got, err := portfolios.List(ctx)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty db: got %+v err %v", got, err)
	}

	for _, name := range []string{"trending", "main"} {
		if err := portfolios.Create(ctx, repository.Portfolio{Name: name, BaseCurrency: "EUR"}); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	trending, _ := portfolios.GetByName(ctx, "trending")
	for _, symbol := range []string{"AAA", "BBB"} {
		inst, err := instruments.GetOrCreate(ctx, repository.Instrument{Symbol: symbol, InstrumentType: "equity", Exchange: "NASDAQ", QuoteCurrency: "USD"})
		if err != nil {
			t.Fatalf("instrument: %v", err)
		}
		err = positions.Create(ctx, repository.CreatePositionParams{InstrumentID: inst.ID, PortfolioID: trending.ID, Quantity: 1, AvgCost: 1, Currency: "USD"})
		if err != nil {
			t.Fatalf("position: %v", err)
		}
	}

	got, err = portfolios.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	want := []repository.PortfolioOverview{
		{Name: "main", BaseCurrency: "EUR", PositionCount: 0},
		{Name: "trending", BaseCurrency: "EUR", PositionCount: 2},
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %+v, want %+v (ordered by name, empty portfolios included)", got, want)
	}
}
