package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/instruments"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase"
)

var errBoom = errors.New("boom")

// world is an in-memory fake of every dependency of the trades import use case.
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

func (w *world) GetByName(_ context.Context, name string) (repository.Portfolio, error) {
	if w.portfolioErr != nil {
		return repository.Portfolio{}, w.portfolioErr
	}
	return w.portfolio, nil
}

func (w *world) GetBySymbolAndExchange(_ context.Context, symbol, exchange string) (repository.Instrument, error) {
	if w.instrumentsErr != nil {
		return repository.Instrument{}, w.instrumentsErr
	}
	inst, ok := w.instruments[symbol+"|"+exchange]
	if !ok {
		return repository.Instrument{}, domain.ErrInstrumentNotFound
	}
	return inst, nil
}

func (w *world) GetOrCreate(_ context.Context, in repository.Instrument) (repository.Instrument, error) {
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

func (w *world) Get(_ context.Context, _, instrumentID int64) (repository.Position, error) {
	if w.getPosErr != nil {
		return repository.Position{}, w.getPosErr
	}
	pos, ok := w.positions[instrumentID]
	if !ok {
		return repository.Position{}, domain.ErrPositionNotFound
	}
	return pos, nil
}

func (w *world) Create(_ context.Context, p repository.CreatePositionParams) error {
	if w.createPosErr != nil {
		return w.createPosErr
	}
	w.positions[p.InstrumentID] = repository.Position{Quantity: p.Quantity, AvgCost: p.AvgCost, Currency: p.Currency, AvgCostBase: p.AvgCostBase}
	return nil
}

func (w *world) Delete(_ context.Context, _, instrumentID int64) error {
	if w.deletePosErr != nil {
		return w.deletePosErr
	}
	delete(w.positions, instrumentID)
	return nil
}

func (w *world) Exists(_ context.Context, _ int64, key string) (bool, error) {
	if w.existsErr != nil {
		return false, w.existsErr
	}
	return w.ledger[key], nil
}

func (w *world) recordTrade(_ context.Context, t repository.Trade) error {
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

func (w *world) GetRate(_ context.Context, from, to string) (domain.FXRate, error) {
	w.fxCalls++
	if w.fxErr != nil {
		return domain.FXRate{}, w.fxErr
	}
	return domain.FXRate{Rate: w.fxRate}, nil
}

// ledgerFake and the positions/ledger split exist because Create is on both interfaces.
type ledgerFake struct{ w *world }

func (l ledgerFake) Exists(ctx context.Context, id int64, key string) (bool, error) {
	return l.w.Exists(ctx, id, key)
}

func (l ledgerFake) Create(ctx context.Context, t repository.Trade) error {
	return l.w.recordTrade(ctx, t)
}

func (w *world) useCase() *usecase.ImportTradesUseCase {
	return usecase.NewImportTradesUseCase(w, w, w, ledgerFake{w}, w, w)
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
		{"instrument create fails", func(w *world) { w.createInstErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, false},
		{"position save fails", func(w *world) { w.createPosErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, false},
		{"position close fails", func(w *world) { seeded(w); w.deletePosErr = errBoom }, input(sell(1, "AAPL", 1, 1)), errBoom, false},
		{"ledger write fails", func(w *world) { w.createTradeErr = errBoom }, input(buy(1, "AAPL", 1, 1)), errBoom, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newWorld()
			if tt.setup != nil {
				tt.setup(w)
			}
			positionsBefore := fmt.Sprint(w.positions)

			out, err := w.useCase().Execute(context.Background(), tt.in)

			if err == nil || out != nil {
				t.Fatalf("expected an error, got out=%+v err=%v", out, err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantNoWrites && (len(w.trades) != 0 || fmt.Sprint(w.positions) != positionsBefore) {
				t.Fatalf("nothing should be written, ledger=%v positions=%v", w.trades, w.positions)
			}
		})
	}
}
