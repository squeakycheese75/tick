package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/instruments"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase"
)

var errBoom = errors.New("boom")

// world is an in-memory fake of every dependency of the trades import use case. It is also the
// unit of work: Within snapshots its state and puts it back if the transaction fails.
type world struct {
	portfolio    repository.Portfolio
	portfolioErr error

	instruments    map[string]repository.Instrument // symbol|exchange
	instrumentsErr error
	createInstErr  error
	nextID         int64

	positions    map[int64]repository.Position
	getPosErr    error
	createPosErr error
	deletePosErr error

	ledger         map[string]bool
	existsErr      error
	createTradeErr error
	trades         []repository.Trade

	resolved    instruments.ResolvedInstrument
	resolverErr error

	fxRate  float64
	fxErr   error
	fxCalls int

	// beforeWrite runs as the write transaction starts, standing in for another writer that
	// got in between planning and writing.
	beforeWrite func(*world)
	commits     int
	rollbacks   int
}

func newWorld() *world {
	return &world{
		portfolio:   repository.Portfolio{ID: 1, Name: "trending", BaseCurrency: "EUR"},
		instruments: map[string]repository.Instrument{},
		positions:   map[int64]repository.Position{},
		ledger:      map[string]bool{},
		nextID:      100,
		resolved:    instruments.ResolvedInstrument{InstrumentType: "equity", Exchange: "NASDAQ", QuoteCurrency: "USD"},
		fxRate:      1.25,
	}
}

type worldState struct {
	instruments map[string]repository.Instrument
	positions   map[int64]repository.Position
	ledger      map[string]bool
	trades      []repository.Trade
	nextID      int64
}

func (w *world) snapshot() worldState {
	st := worldState{
		instruments: map[string]repository.Instrument{},
		positions:   map[int64]repository.Position{},
		ledger:      map[string]bool{},
		trades:      append([]repository.Trade(nil), w.trades...),
		nextID:      w.nextID,
	}
	for k, v := range w.instruments {
		st.instruments[k] = v
	}
	for k, v := range w.positions {
		st.positions[k] = v
	}
	for k, v := range w.ledger {
		st.ledger[k] = v
	}

	return st
}

func (w *world) restore(st worldState) {
	w.instruments, w.positions, w.ledger, w.trades, w.nextID = st.instruments, st.positions, st.ledger, st.trades, st.nextID
}

// Reader reads the world as it is.
func (w *world) Reader() usecase.TradeStore { return w }

// Within runs fn as one transaction: kept if it succeeds, undone if it fails.
func (w *world) Within(_ context.Context, fn func(usecase.TradeStore) error) error {
	if w.beforeWrite != nil {
		w.beforeWrite(w)
	}

	before := w.snapshot()

	if err := fn(w); err != nil {
		w.restore(before)
		w.rollbacks++
		return err
	}

	w.commits++

	return nil
}

func (w *world) GetPortfolio(_ context.Context, _ string) (repository.Portfolio, error) {
	if w.portfolioErr != nil {
		return repository.Portfolio{}, w.portfolioErr
	}
	return w.portfolio, nil
}

func (w *world) GetInstrument(_ context.Context, symbol, exchange string) (repository.Instrument, error) {
	if w.instrumentsErr != nil {
		return repository.Instrument{}, w.instrumentsErr
	}
	inst, ok := w.instruments[symbol+"|"+exchange]
	if !ok {
		return repository.Instrument{}, domain.ErrInstrumentNotFound
	}
	return inst, nil
}

func (w *world) GetOrCreateInstrument(_ context.Context, in repository.Instrument) (repository.Instrument, error) {
	if w.createInstErr != nil {
		return repository.Instrument{}, w.createInstErr
	}
	key := in.Symbol + "|" + in.Exchange
	if inst, ok := w.instruments[key]; ok {
		return inst, nil
	}
	w.nextID++
	in.ID = w.nextID
	w.instruments[key] = in
	return in, nil
}

func (w *world) GetPosition(_ context.Context, _, instrumentID int64) (repository.Position, error) {
	if w.getPosErr != nil {
		return repository.Position{}, w.getPosErr
	}
	pos, ok := w.positions[instrumentID]
	if !ok {
		return repository.Position{}, domain.ErrPositionNotFound
	}
	return pos, nil
}

func (w *world) SavePosition(_ context.Context, p repository.CreatePositionParams) error {
	if w.createPosErr != nil {
		return w.createPosErr
	}
	w.positions[p.InstrumentID] = repository.Position{Quantity: p.Quantity, AvgCost: p.AvgCost, Currency: p.Currency, AvgCostBase: p.AvgCostBase}
	return nil
}

func (w *world) DeletePosition(_ context.Context, _, instrumentID int64) error {
	if w.deletePosErr != nil {
		return w.deletePosErr
	}
	delete(w.positions, instrumentID)
	return nil
}

func (w *world) TradeRecorded(_ context.Context, _ int64, key string) (bool, error) {
	if w.existsErr != nil {
		return false, w.existsErr
	}
	return w.ledger[key], nil
}

func (w *world) RecordTrade(_ context.Context, t repository.Trade) error {
	if w.createTradeErr != nil {
		return w.createTradeErr
	}
	w.ledger[t.Key] = true
	w.trades = append(w.trades, t)
	return nil
}

func (w *world) Resolve(_ context.Context, _ string) (instruments.ResolvedInstrument, error) {
	return w.resolved, w.resolverErr
}

func (w *world) GetRate(_ context.Context, _, _ string) (domain.FXRate, error) {
	w.fxCalls++
	if w.fxErr != nil {
		return domain.FXRate{}, w.fxErr
	}
	return domain.FXRate{Rate: w.fxRate}, nil
}

func (w *world) useCase() *usecase.ImportTradesUseCase {
	return usecase.NewImportTradesUseCase(w, w, w)
}

var day = func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

func buy(d int, symbol string, qty, price float64) domain.TradeInput {
	return domain.TradeInput{
		Date: day(d), Symbol: symbol, Side: domain.TradeSideBuy,
		Quantity: qty, Price: price, Currency: "USD",
	}
}

func sell(d int, symbol string, qty, price float64) domain.TradeInput {
	t := buy(d, symbol, qty, price)
	t.Side = domain.TradeSideSell
	return t
}

func input(trades ...domain.TradeInput) domain.ImportTradesInput {
	return domain.ImportTradesInput{PortfolioName: "trending", Trades: trades}
}

func (w *world) position(t *testing.T, symbol string) (repository.Position, bool) {
	t.Helper()
	inst, ok := w.instruments[symbol+"|NASDAQ"]
	if !ok {
		return repository.Position{}, false
	}
	pos, ok := w.positions[inst.ID]
	return pos, ok
}

func TestImportTrades_Execute_Success(t *testing.T) {
	t.Parallel()

	eurBuy := buy(1, "AAPL", 10, 100)
	eurBuy.Currency = "EUR"
	eurBuyFee := eurBuy
	eurBuyFee.Fee = 8

	explicitFX := eurBuy
	explicitFX.FXRate = 2

	tests := []struct {
		name        string
		setup       func(*world)
		trades      []domain.TradeInput
		wantQty     float64
		wantCost    float64
		wantBase    float64
		wantOpen    bool
		wantApplied int
		wantSkipped int
		wantFXCalls int
	}{
		{
			name:   "buy opens a position",
			trades: []domain.TradeInput{buy(1, "AAPL", 10, 100)}, wantQty: 10, wantCost: 100, wantBase: 125, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
		{
			name:   "trade currency converted to instrument currency at the provider rate",
			trades: []domain.TradeInput{eurBuy}, wantQty: 10, wantCost: 125, wantBase: 100, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
		{
			name:   "fee is converted and spread into cost",
			trades: []domain.TradeInput{eurBuyFee}, wantQty: 10, wantCost: 126, wantBase: 100.8, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
		{
			name:   "explicit fx rate wins over the provider",
			trades: []domain.TradeInput{explicitFX}, wantQty: 10, wantCost: 200, wantBase: 100, wantOpen: true, wantApplied: 1,
		},
		{
			name:   "buys average together",
			trades: []domain.TradeInput{buy(1, "AAPL", 10, 100), buy(2, "AAPL", 10, 200)}, wantQty: 20, wantCost: 150, wantBase: 187.5, wantOpen: true, wantApplied: 2, wantFXCalls: 2,
		},
		{
			name:   "trades are applied in date order regardless of file order",
			trades: []domain.TradeInput{sell(3, "AAPL", 4, 120), buy(1, "AAPL", 10, 100)}, wantQty: 6, wantCost: 100, wantBase: 125, wantOpen: true, wantApplied: 2, wantFXCalls: 1,
		},
		{
			name:   "selling everything closes the position",
			trades: []domain.TradeInput{buy(1, "AAPL", 10, 100), sell(2, "AAPL", 10, 120)}, wantOpen: false, wantApplied: 2, wantFXCalls: 1,
		},
		{
			name: "buy adds to an existing position",
			setup: func(w *world) {
				w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "USD"}
				w.positions[7] = repository.Position{Quantity: 10, AvgCost: 100}
			},
			trades: []domain.TradeInput{buy(1, "AAPL", 10, 200)}, wantQty: 20, wantCost: 150, wantBase: 0, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
		{
			name: "known instrument without a position starts from nothing",
			setup: func(w *world) {
				w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "USD"}
			},
			trades: []domain.TradeInput{buy(1, "AAPL", 2, 50)}, wantQty: 2, wantCost: 50, wantBase: 62.5, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
		{
			name: "stored instrument currency wins over the trade's",
			setup: func(w *world) {
				w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "EUR"}
			},
			trades: []domain.TradeInput{eurBuy}, wantQty: 10, wantCost: 100, wantBase: 100, wantOpen: true, wantApplied: 1,
		},
		{
			name: "existing position with a known base cost keeps averaging it",
			setup: func(w *world) {
				w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "USD"}
				w.positions[7] = repository.Position{Quantity: 10, AvgCost: 100, AvgCostBase: 80}
			},
			trades: []domain.TradeInput{buy(1, "AAPL", 10, 200)}, wantQty: 20, wantCost: 150, wantBase: 165, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
		{
			name: "trade priced in a currency that is not the base uses the explicit rate when the instrument is in the base currency",
			setup: func(w *world) {
				w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "EUR"}
			},
			trades: []domain.TradeInput{func() domain.TradeInput {
				tr := buy(1, "AAPL", 10, 100)
				tr.FXRate = 0.9
				return tr
			}()},
			wantQty: 10, wantCost: 90, wantBase: 90, wantOpen: true, wantApplied: 1,
		},
		{
			name: "trades already in the ledger are skipped",
			setup: func(w *world) {
				w.ledger["2026-10-01|AAPL|buy|10|100|USD#1"] = true
			},
			trades: []domain.TradeInput{buy(1, "AAPL", 10, 100)}, wantOpen: false, wantSkipped: 1,
		},
		{
			name:   "identical fills in one file are distinct trades",
			trades: []domain.TradeInput{buy(1, "AAPL", 1, 100), buy(1, "AAPL", 1, 100)}, wantQty: 2, wantCost: 100, wantBase: 125, wantOpen: true, wantApplied: 2, wantFXCalls: 2,
		},
		{
			name: "a ref identifies the trade",
			trades: []domain.TradeInput{func() domain.TradeInput {
				tr := buy(1, "AAPL", 1, 100)
				tr.Ref = "broker-1"
				return tr
			}()},
			setup:    func(w *world) { w.ledger["broker-1"] = true },
			wantOpen: false, wantSkipped: 1,
		},
		{
			name: "explicit instrument metadata skips the resolver",
			setup: func(w *world) {
				w.resolverErr = errBoom
			},
			trades: []domain.TradeInput{func() domain.TradeInput {
				tr := buy(1, "AAPL", 1, 100)
				tr.InstrumentType, tr.Exchange, tr.QuoteCurrency = "equity", "NASDAQ", "USD"
				return tr
			}()},
			wantQty: 1, wantCost: 100, wantBase: 125, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
		{
			name: "partly explicit metadata is completed by the resolver",
			setup: func(w *world) {
				w.resolved = instruments.ResolvedInstrument{InstrumentType: "equity", Exchange: "NYSE", QuoteCurrency: "USD"}
			},
			trades: []domain.TradeInput{func() domain.TradeInput {
				tr := buy(1, "AAPL", 1, 100)
				tr.Exchange = "NASDAQ"
				return tr
			}()},
			wantQty: 1, wantCost: 100, wantBase: 125, wantOpen: true, wantApplied: 1, wantFXCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newWorld()
			if tt.setup != nil {
				tt.setup(w)
			}

			out, err := w.useCase().Execute(context.Background(), input(tt.trades...))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if out.Applied != tt.wantApplied || out.Skipped != tt.wantSkipped || out.PortfolioName != "trending" {
				t.Fatalf("output = %+v, want applied %d skipped %d", out, tt.wantApplied, tt.wantSkipped)
			}
			if len(w.trades) != tt.wantApplied {
				t.Fatalf("ledger rows = %d, want %d", len(w.trades), tt.wantApplied)
			}
			if w.fxCalls != tt.wantFXCalls {
				t.Fatalf("fx calls = %d, want %d", w.fxCalls, tt.wantFXCalls)
			}

			pos, open := w.position(t, "AAPL")
			if open != tt.wantOpen {
				t.Fatalf("position open = %v, want %v", open, tt.wantOpen)
			}
			if open && (pos.Quantity != tt.wantQty || pos.AvgCost != tt.wantCost || math.Abs(pos.AvgCostBase-tt.wantBase) > 1e-9) {
				t.Fatalf("position = %+v, want qty %v cost %v base %v", pos, tt.wantQty, tt.wantCost, tt.wantBase)
			}
		})
	}
}

func TestImportTrades_Execute_RecordsLedgerRow(t *testing.T) {
	t.Parallel()

	w := newWorld()
	tr := buy(1, "AAPL", 10, 100)
	tr.Currency = "EUR"
	tr.Fee = 2

	if _, err := w.useCase().Execute(context.Background(), input(tr)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := repository.Trade{
		PortfolioID: 1, InstrumentID: 101, Key: "2026-10-01|AAPL|buy|10|100|EUR#1",
		Side: "buy", Quantity: 10, Price: 100, Currency: "EUR", Fee: 2, FXRate: 1.25, TradedAt: day(1),
	}
	if len(w.trades) != 1 || w.trades[0] != want {
		t.Fatalf("ledger = %+v, want %+v", w.trades, want)
	}
}

func TestImportTrades_Execute_Failures(t *testing.T) {
	t.Parallel()

	seeded := func(w *world) {
		w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "USD"}
		w.positions[7] = repository.Position{Quantity: 1, AvgCost: 100}
	}
	eur := buy(1, "AAPL", 1, 100)
	eur.Currency = "EUR"

	tests := []struct {
		name         string
		setup        func(*world)
		in           domain.ImportTradesInput
		wantErr      error
		wantNoWrites bool
	}{
		{"invalid input", nil, domain.ImportTradesInput{PortfolioName: "trending"}, nil, true},
		{"portfolio missing", func(w *world) { w.portfolioErr = domain.ErrPortfolioNotFound }, input(buy(1, "AAPL", 1, 1)), domain.ErrPortfolioNotFound, true},
		{"portfolio lookup fails", func(w *world) { w.portfolioErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
		{"ledger check fails", func(w *world) { w.existsErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
		{"metadata cannot be resolved", func(w *world) { w.resolverErr = errBoom }, input(buy(1, "AAPL", 1, 1)), nil, true},
		{"fx lookup fails", func(w *world) { w.fxErr = errBoom }, input(eur), errBoom, true},
		{"base currency rate lookup fails", func(w *world) { w.fxErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
		{"instrument lookup fails", func(w *world) { w.instrumentsErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
		{"position lookup fails", func(w *world) { seeded(w); w.getPosErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
		{"overselling is rejected", seeded, input(sell(1, "AAPL", 5, 1)), domain.ErrInsufficientQuantity, true},
		{
			"a bad row later in the file stops everything",
			nil, input(buy(1, "AAPL", 1, 1), sell(2, "AAPL", 5, 1)), domain.ErrInsufficientQuantity, true,
		},
		{"instrument create fails", func(w *world) { w.createInstErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
		{"position save fails", func(w *world) { w.createPosErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
		{"position close fails", func(w *world) { seeded(w); w.deletePosErr = errBoom }, input(sell(1, "AAPL", 1, 1)), errBoom, true},
		{"ledger write fails", func(w *world) { w.createTradeErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newWorld()
			if tt.setup != nil {
				tt.setup(w)
			}
			before := fmt.Sprint(w.instruments, w.positions, w.ledger, w.trades)

			out, err := w.useCase().Execute(context.Background(), tt.in)

			if err == nil || out != nil {
				t.Fatalf("expected an error, got out=%+v err=%v", out, err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantNoWrites && fmt.Sprint(w.instruments, w.positions, w.ledger, w.trades) != before {
				t.Fatalf("a failed import must leave nothing behind: instruments=%v positions=%v trades=%v", w.instruments, w.positions, w.trades)
			}
			if w.commits != 0 {
				t.Fatalf("a failed import must not commit, got %d commits", w.commits)
			}
		})
	}
}

func TestImportTrades_Execute_FailureWhileWritingIsRolledBack(t *testing.T) {
	t.Parallel()

	// The first trade writes fine; the second fails on the ledger. The first must not survive.
	w := newWorld()
	w.createTradeErr = errBoom

	_, err := w.useCase().Execute(context.Background(), input(buy(1, "AAPL", 1, 100), buy(2, "MSFT", 1, 100)))

	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if w.rollbacks != 1 || w.commits != 0 {
		t.Fatalf("rollbacks = %d commits = %d, want 1 and 0", w.rollbacks, w.commits)
	}
	if len(w.positions) != 0 || len(w.instruments) != 0 {
		t.Fatalf("partial import survived: positions=%v instruments=%v", w.positions, w.instruments)
	}
}

func TestImportTrades_Execute_PlansAgainstTheStateItHoldsWhenWriting(t *testing.T) {
	t.Parallel()

	// Another writer buys 10 @ 100 after the first plan but before this import's write begins.
	// Planning from the first look would give 10 @ 100; planning inside the transaction sees
	// the other writer's shares and averages with them.
	w := newWorld()
	w.beforeWrite = func(w *world) {
		w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "USD"}
		w.positions[7] = repository.Position{Quantity: 10, AvgCost: 100, AvgCostBase: 80}
	}

	if _, err := w.useCase().Execute(context.Background(), input(buy(1, "AAPL", 10, 200))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pos := w.positions[7]
	if pos.Quantity != 20 || pos.AvgCost != 150 {
		t.Fatalf("position = %+v, want 20 @ 150 (the other writer's shares must not be lost)", pos)
	}
}

func TestImportTrades_Execute_StopsWhenTheWriteNeedsARateItNeverFetched(t *testing.T) {
	t.Parallel()

	// At first AAPL is unknown and quoted in USD, so no rate is needed for a USD trade into a
	// USD instrument... but another writer then makes it a EUR instrument, which needs USD/EUR.
	// The transaction must not call out for it (the provider's cache needs the write lock the
	// transaction holds), so it stops and says to run it again.
	w := newWorld()
	w.portfolio.BaseCurrency = "USD"
	w.beforeWrite = func(w *world) {
		w.instruments["AAPL|NASDAQ"] = repository.Instrument{ID: 7, Symbol: "AAPL", Exchange: "NASDAQ", QuoteCurrency: "EUR"}
	}

	callsBefore := w.fxCalls
	_, err := w.useCase().Execute(context.Background(), input(buy(1, "AAPL", 1, 100)))

	if err == nil || !strings.Contains(err.Error(), "run it again") {
		t.Fatalf("err = %v, want a message saying to run it again", err)
	}
	if w.fxCalls != callsBefore {
		t.Fatalf("the transaction called the rate provider %d times", w.fxCalls-callsBefore)
	}
	if w.rollbacks != 1 || len(w.trades) != 0 {
		t.Fatalf("rollbacks = %d trades = %v", w.rollbacks, w.trades)
	}
}

func TestImportTrades_Execute_RatesAreFetchedBeforeTheTransactionOnly(t *testing.T) {
	t.Parallel()

	w := newWorld()
	eur := buy(1, "AAPL", 1, 100)
	eur.Currency = "EUR"

	if _, err := w.useCase().Execute(context.Background(), input(eur)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.fxCalls != 1 {
		t.Fatalf("fx calls = %d, want exactly 1 (fetched while planning, replayed while writing)", w.fxCalls)
	}
	if w.commits != 1 {
		t.Fatalf("commits = %d, want 1", w.commits)
	}
}
