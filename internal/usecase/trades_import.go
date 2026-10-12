package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type (
	// TradeStore is everything importing trades reads and writes.
	TradeStore interface {
		GetPortfolio(ctx context.Context, name string) (repository.Portfolio, error)
		GetInstrument(ctx context.Context, symbol, exchange string) (repository.Instrument, error)
		GetOrCreateInstrument(ctx context.Context, in repository.Instrument) (repository.Instrument, error)
		GetPosition(ctx context.Context, portfolioID, instrumentID int64) (repository.Position, error)
		SavePosition(ctx context.Context, p repository.CreatePositionParams) error
		DeletePosition(ctx context.Context, portfolioID, instrumentID int64) error
		TradeRecorded(ctx context.Context, portfolioID int64, key string) (bool, error)
		RecordTrade(ctx context.Context, t repository.Trade) error
	}
	// TradeUnitOfWork gives a TradeStore that reads freely, and runs groups of operations
	// atomically: all of them are kept or none are.
	TradeUnitOfWork interface {
		Reader() TradeStore
		Within(ctx context.Context, fn func(TradeStore) error) error
	}
	tradeFXRates interface {
		GetRate(ctx context.Context, from, to string) (domain.FXRate, error)
	}
)

// ImportTradesUseCase applies a batch of executed trades to a portfolio's positions
// and records them in the trade ledger.
type ImportTradesUseCase struct {
	uow      TradeUnitOfWork
	resolver InstrumentResolver
	fx       tradeFXRates
}

// NewImportTradesUseCase builds an ImportTradesUseCase.
func NewImportTradesUseCase(
	uow TradeUnitOfWork,
	resolver InstrumentResolver,
	fx tradeFXRates,
) *ImportTradesUseCase {
	return &ImportTradesUseCase{uow: uow, resolver: resolver, fx: fx}
}

// plannedTrade is a validated trade together with the position it leaves behind.
type plannedTrade struct {
	trade      domain.TradeInput
	key        string
	instrument repository.Instrument
	fxRate     float64
	after      domain.TradePosition
}

// tracked is an instrument and the position held in it as the plan is built up.
type tracked struct {
	instrument repository.Instrument
	position   domain.TradePosition
}

// tradePlan is the outcome of planning a batch against some state of the store.
type tradePlan struct {
	portfolio repository.Portfolio
	trades    []plannedTrade
	skipped   int
}

// Execute imports the trades in two phases. First it plans the whole batch against the
// current state without holding any lock, which fetches every exchange rate it needs and
// rejects a bad row before anything is written. Then, in one write transaction, it plans
// again against the state it now holds exclusively, from the rates already fetched, and
// writes the result. Nothing can change between that second plan and the write, and a
// failure at any point leaves nothing behind. Trades already in the ledger are skipped.
func (uc *ImportTradesUseCase) Execute(
	ctx context.Context,
	in domain.ImportTradesInput,
) (*domain.ImportTradesOutput, error) {
	for i := range in.Trades {
		in.Trades[i].Normalize()
	}

	if err := in.Validate(); err != nil {
		return nil, err
	}

	rates := &recordedRates{live: uc.fx, seen: map[string]domain.FXRate{}}

	if _, err := uc.plan(ctx, uc.uow.Reader(), in, rates.fetching()); err != nil {
		return nil, err
	}

	var out *domain.ImportTradesOutput

	err := uc.uow.Within(ctx, func(store TradeStore) error {
		plan, err := uc.plan(ctx, store, in, rates.replaying())
		if err != nil {
			return err
		}

		for _, p := range plan.trades {
			if err := uc.apply(ctx, store, plan.portfolio.ID, p); err != nil {
				return err
			}
		}

		out = &domain.ImportTradesOutput{
			PortfolioName: plan.portfolio.Name,
			Applied:       len(plan.trades),
			Skipped:       plan.skipped,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (uc *ImportTradesUseCase) plan(
	ctx context.Context,
	store TradeStore,
	in domain.ImportTradesInput,
	rates tradeFXRates,
) (tradePlan, error) {
	portfolio, err := store.GetPortfolio(ctx, in.PortfolioName)
	if errors.Is(err, domain.ErrPortfolioNotFound) {
		return tradePlan{}, fmt.Errorf("%w: %q", domain.ErrPortfolioNotFound, in.PortfolioName)
	}
	if err != nil {
		return tradePlan{}, fmt.Errorf("get portfolio %q: %w", in.PortfolioName, err)
	}

	ordered := append([]domain.TradeInput(nil), in.Trades...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Date.Before(ordered[j].Date) })

	keys := tradeKeys(ordered)
	state := map[string]tracked{}
	plan := tradePlan{portfolio: portfolio, trades: make([]plannedTrade, 0, len(ordered))}

	for i, t := range ordered {
		recorded, err := store.TradeRecorded(ctx, portfolio.ID, keys[i])
		if err != nil {
			return tradePlan{}, err
		}
		if recorded {
			plan.skipped++
			continue
		}

		p, err := uc.planTrade(ctx, store, portfolio, t, keys[i], state, rates)
		if err != nil {
			return tradePlan{}, fmt.Errorf("%s %s on %s: %w", t.Side, t.Symbol, t.Date.Format("2006-01-02"), err)
		}

		plan.trades = append(plan.trades, p)
	}

	return plan, nil
}

func (uc *ImportTradesUseCase) planTrade(
	ctx context.Context,
	store TradeStore,
	portfolio repository.Portfolio,
	t domain.TradeInput,
	key string,
	state map[string]tracked,
	rates tradeFXRates,
) (plannedTrade, error) {
	t, err := uc.resolveMetadata(ctx, t)
	if err != nil {
		return plannedTrade{}, err
	}

	stateKey := t.Symbol + "|" + t.Exchange

	cur, seen := state[stateKey]
	if !seen {
		cur, err = load(ctx, store, portfolio.ID, t)
		if err != nil {
			return plannedTrade{}, err
		}
	}

	t.QuoteCurrency = cur.instrument.QuoteCurrency

	fxRate, err := tradeToQuoteRate(ctx, rates, t)
	if err != nil {
		return plannedTrade{}, err
	}

	after, err := applyTrade(ctx, rates, cur.position, t, fxRate, portfolio.BaseCurrency)
	if err != nil {
		return plannedTrade{}, err
	}

	state[stateKey] = tracked{instrument: cur.instrument, position: after}

	return plannedTrade{trade: t, key: key, instrument: cur.instrument, fxRate: fxRate, after: after}, nil
}

func applyTrade(
	ctx context.Context,
	rates tradeFXRates,
	current domain.TradePosition,
	t domain.TradeInput,
	fxRate float64,
	baseCurrency string,
) (domain.TradePosition, error) {
	if t.Side == domain.TradeSideSell {
		return current.ApplySell(t.Quantity)
	}

	baseRate, err := tradeToBaseRate(ctx, rates, t, fxRate, baseCurrency)
	if err != nil {
		return domain.TradePosition{}, err
	}

	gross := t.Quantity*t.Price + t.Fee

	return current.ApplyBuy(t.Quantity, gross*fxRate, gross*baseRate), nil
}

// tradeToBaseRate converts the trade currency to the portfolio base currency, so the cost in base
// currency is recorded as it was when bought. Only a trade in a different currency needs a lookup.
func tradeToBaseRate(
	ctx context.Context,
	rates tradeFXRates,
	t domain.TradeInput,
	fxRate float64,
	baseCurrency string,
) (float64, error) {
	if t.Currency == baseCurrency {
		return 1, nil
	}
	if t.QuoteCurrency == baseCurrency {
		return fxRate, nil
	}

	rate, err := rates.GetRate(ctx, t.Currency, baseCurrency)
	if err != nil {
		return 0, fmt.Errorf("%w: fx rate %s to %s: %w", domain.ErrUpstream, t.Currency, baseCurrency, err)
	}

	return rate.Rate, nil
}

// load returns the instrument and what is held of it. An unknown instrument or position means
// nothing is held; a known instrument's stored identity replaces the one derived from the trade.
func load(
	ctx context.Context,
	store TradeStore,
	portfolioID int64,
	t domain.TradeInput,
) (tracked, error) {
	fresh := tracked{instrument: repository.Instrument{
		Symbol:         t.Symbol,
		ProviderSymbol: t.Symbol,
		InstrumentType: t.InstrumentType,
		Exchange:       t.Exchange,
		QuoteCurrency:  t.QuoteCurrency,
	}}

	existing, err := store.GetInstrument(ctx, t.Symbol, t.Exchange)
	if errors.Is(err, domain.ErrInstrumentNotFound) {
		return fresh, nil
	}
	if err != nil {
		return tracked{}, err
	}

	pos, err := store.GetPosition(ctx, portfolioID, existing.ID)
	if errors.Is(err, domain.ErrPositionNotFound) {
		return tracked{instrument: existing}, nil
	}
	if err != nil {
		return tracked{}, err
	}

	return tracked{
		instrument: existing,
		position:   domain.TradePosition{Quantity: pos.Quantity, AvgCost: pos.AvgCost, AvgCostBase: pos.AvgCostBase},
	}, nil
}

func (uc *ImportTradesUseCase) resolveMetadata(
	ctx context.Context,
	t domain.TradeInput,
) (domain.TradeInput, error) {
	if t.InstrumentType != "" && t.Exchange != "" && t.QuoteCurrency != "" {
		return t, nil
	}

	resolved, err := uc.resolver.Resolve(ctx, t.Symbol)
	if err != nil {
		return t, domain.Invalid(
			"could not resolve instrument metadata for %q; give its instrument_type, exchange and quote_currency",
			t.Symbol,
		)
	}

	if t.InstrumentType == "" {
		t.InstrumentType = resolved.InstrumentType
	}
	if t.Exchange == "" {
		t.Exchange = resolved.Exchange
	}
	if t.QuoteCurrency == "" {
		t.QuoteCurrency = resolved.QuoteCurrency
	}

	return t, nil
}

// tradeToQuoteRate converts the trade currency to the instrument's quote currency.
// An explicit rate on the trade wins; otherwise the current rate is used.
func tradeToQuoteRate(ctx context.Context, rates tradeFXRates, t domain.TradeInput) (float64, error) {
	if t.Currency == t.QuoteCurrency {
		return 1, nil
	}
	if t.FXRate > 0 {
		return t.FXRate, nil
	}

	rate, err := rates.GetRate(ctx, t.Currency, t.QuoteCurrency)
	if err != nil {
		return 0, fmt.Errorf("%w: fx rate %s to %s: %w", domain.ErrUpstream, t.Currency, t.QuoteCurrency, err)
	}

	return rate.Rate, nil
}

func (uc *ImportTradesUseCase) apply(ctx context.Context, store TradeStore, portfolioID int64, p plannedTrade) error {
	instrument, err := store.GetOrCreateInstrument(ctx, p.instrument)
	if err != nil {
		return fmt.Errorf("get or create instrument %q: %w", p.trade.Symbol, err)
	}

	if err := savePosition(ctx, store, portfolioID, instrument, p.after); err != nil {
		return err
	}

	return store.RecordTrade(ctx, repository.Trade{
		PortfolioID:  portfolioID,
		InstrumentID: instrument.ID,
		Key:          p.key,
		Side:         string(p.trade.Side),
		Quantity:     p.trade.Quantity,
		Price:        p.trade.Price,
		Currency:     p.trade.Currency,
		Fee:          p.trade.Fee,
		FXRate:       p.fxRate,
		TradedAt:     p.trade.Date,
	})
}

func savePosition(
	ctx context.Context,
	store TradeStore,
	portfolioID int64,
	instrument repository.Instrument,
	after domain.TradePosition,
) error {
	if after.IsClosed() {
		if err := store.DeletePosition(ctx, portfolioID, instrument.ID); err != nil {
			return fmt.Errorf("close position %q: %w", instrument.Symbol, err)
		}
		return nil
	}

	err := store.SavePosition(ctx, repository.CreatePositionParams{
		InstrumentID: instrument.ID,
		PortfolioID:  portfolioID,
		Quantity:     after.Quantity,
		AvgCost:      after.AvgCost,
		AvgCostBase:  after.AvgCostBase,
		Currency:     instrument.QuoteCurrency,
	})
	if err != nil {
		return fmt.Errorf("save position %q: %w", instrument.Symbol, err)
	}

	return nil
}

// tradeKeys gives each trade a stable identity: its own ref, or a fingerprint of its fields
// plus an occurrence number so genuine repeats of the same fill within a file stay distinct.
func tradeKeys(trades []domain.TradeInput) []string {
	seen := map[string]int{}
	keys := make([]string, len(trades))

	for i, t := range trades {
		if t.Ref != "" {
			keys[i] = t.Ref
			continue
		}

		base := fmt.Sprintf(
			"%s|%s|%s|%g|%g|%s",
			t.Date.Format("2006-01-02"),
			t.Symbol,
			t.Side,
			t.Quantity,
			t.Price,
			t.Currency,
		)
		seen[base]++
		keys[i] = fmt.Sprintf("%s#%d", base, seen[base])
	}

	return keys
}

// recordedRates fetches exchange rates from the live source and remembers them, so the plan
// made inside the write transaction can use exactly those rates without calling out again:
// the provider's cache is written through another connection, which would wait on the very
// lock the transaction holds.
type recordedRates struct {
	live tradeFXRates
	mu   sync.Mutex
	seen map[string]domain.FXRate
}

// fetching returns a rate source that calls the live provider and records each answer.
func (r *recordedRates) fetching() tradeFXRates { return fetchingRates{r} }

// replaying returns a rate source that only answers with what was recorded.
func (r *recordedRates) replaying() tradeFXRates { return replayingRates{r} }

type fetchingRates struct{ r *recordedRates }

func (f fetchingRates) GetRate(ctx context.Context, from, to string) (domain.FXRate, error) {
	rate, err := f.r.live.GetRate(ctx, from, to)
	if err != nil {
		return domain.FXRate{}, err
	}

	f.r.mu.Lock()
	f.r.seen[from+"/"+to] = rate
	f.r.mu.Unlock()

	return rate, nil
}

type replayingRates struct{ r *recordedRates }

func (p replayingRates) GetRate(_ context.Context, from, to string) (domain.FXRate, error) {
	p.r.mu.Lock()
	defer p.r.mu.Unlock()

	rate, ok := p.r.seen[from+"/"+to]
	if !ok {
		return domain.FXRate{}, fmt.Errorf(
			"%w: exchange rate %s to %s was not fetched before writing, so run it again",
			domain.ErrConcurrentChange,
			from,
			to,
		)
	}

	return rate, nil
}
