package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type (
	tradePortfolios interface {
		GetByName(ctx context.Context, name string) (repository.Portfolio, error)
	}
	tradeInstruments interface {
		GetBySymbolAndExchange(ctx context.Context, symbol, exchange string) (repository.Instrument, error)
		GetOrCreate(ctx context.Context, in repository.Instrument) (repository.Instrument, error)
	}
	tradePositions interface {
		Get(ctx context.Context, portfolioID, instrumentID int64) (repository.Position, error)
		Create(ctx context.Context, p repository.CreatePositionParams) error
		Delete(ctx context.Context, portfolioID, instrumentID int64) error
	}
	tradeLedger interface {
		Exists(ctx context.Context, portfolioID int64, key string) (bool, error)
		Create(ctx context.Context, t repository.Trade) error
	}
	tradeFXRates interface {
		GetRate(ctx context.Context, from, to string) (domain.FXRate, error)
	}
)

// ImportTradesUseCase applies a batch of executed trades to a portfolio's positions
// and records them in the trade ledger.
type ImportTradesUseCase struct {
	portfolios  tradePortfolios
	instruments tradeInstruments
	positions   tradePositions
	ledger      tradeLedger
	resolver    InstrumentResolver
	fx          tradeFXRates
}

// NewImportTradesUseCase builds an ImportTradesUseCase.
func NewImportTradesUseCase(
	portfolios tradePortfolios,
	instruments tradeInstruments,
	positions tradePositions,
	ledger tradeLedger,
	resolver InstrumentResolver,
	fx tradeFXRates,
) *ImportTradesUseCase {
	return &ImportTradesUseCase{
		portfolios:  portfolios,
		instruments: instruments,
		positions:   positions,
		ledger:      ledger,
		resolver:    resolver,
		fx:          fx,
	}
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

// Execute validates and plans every trade first, so a bad row fails the whole file
// before anything is written. Trades already in the ledger are skipped.
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

	portfolio, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return nil, fmt.Errorf("%w: %q", domain.ErrPortfolioNotFound, in.PortfolioName)
		}
		return nil, fmt.Errorf("get portfolio %q: %w", in.PortfolioName, err)
	}

	plan, skipped, err := uc.plan(ctx, portfolio, in.Trades)
	if err != nil {
		return nil, err
	}

	for _, p := range plan {
		if err := uc.apply(ctx, portfolio.ID, p); err != nil {
			return nil, err
		}
	}

	return &domain.ImportTradesOutput{
		PortfolioName: portfolio.Name,
		Applied:       len(plan),
		Skipped:       skipped,
	}, nil
}

func (uc *ImportTradesUseCase) plan(
	ctx context.Context,
	portfolio repository.Portfolio,
	trades []domain.TradeInput,
) ([]plannedTrade, int, error) {
	ordered := append([]domain.TradeInput(nil), trades...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Date.Before(ordered[j].Date) })

	keys := tradeKeys(ordered)
	state := map[string]tracked{}
	plan := make([]plannedTrade, 0, len(ordered))
	skipped := 0

	for i, t := range ordered {
		exists, err := uc.ledger.Exists(ctx, portfolio.ID, keys[i])
		if err != nil {
			return nil, 0, err
		}
		if exists {
			skipped++
			continue
		}

		p, err := uc.planTrade(ctx, portfolio, t, keys[i], state)
		if err != nil {
			return nil, 0, fmt.Errorf("%s %s on %s: %w", t.Side, t.Symbol, t.Date.Format("2006-01-02"), err)
		}

		plan = append(plan, p)
	}

	return plan, skipped, nil
}

func (uc *ImportTradesUseCase) planTrade(
	ctx context.Context,
	portfolio repository.Portfolio,
	t domain.TradeInput,
	key string,
	state map[string]tracked,
) (plannedTrade, error) {
	t, err := uc.resolveMetadata(ctx, t)
	if err != nil {
		return plannedTrade{}, err
	}

	stateKey := t.Symbol + "|" + t.Exchange

	cur, seen := state[stateKey]
	if !seen {
		cur, err = uc.load(ctx, portfolio.ID, t)
		if err != nil {
			return plannedTrade{}, err
		}
	}

	t.QuoteCurrency = cur.instrument.QuoteCurrency

	fxRate, err := uc.fxRate(ctx, t)
	if err != nil {
		return plannedTrade{}, err
	}

	after, err := uc.applyTrade(ctx, cur.position, t, fxRate, portfolio.BaseCurrency)
	if err != nil {
		return plannedTrade{}, err
	}

	state[stateKey] = tracked{instrument: cur.instrument, position: after}

	return plannedTrade{trade: t, key: key, instrument: cur.instrument, fxRate: fxRate, after: after}, nil
}

func (uc *ImportTradesUseCase) applyTrade(
	ctx context.Context,
	current domain.TradePosition,
	t domain.TradeInput,
	fxRate float64,
	baseCurrency string,
) (domain.TradePosition, error) {
	if t.Side == domain.TradeSideSell {
		return current.ApplySell(t.Quantity)
	}

	baseRate, err := uc.baseRate(ctx, t, fxRate, baseCurrency)
	if err != nil {
		return domain.TradePosition{}, err
	}

	gross := t.Quantity*t.Price + t.Fee

	return current.ApplyBuy(t.Quantity, gross*fxRate, gross*baseRate), nil
}

// baseRate converts the trade currency to the portfolio base currency, so the cost in base
// currency is recorded as it was when bought. Only a trade in a different currency needs a lookup.
func (uc *ImportTradesUseCase) baseRate(
	ctx context.Context,
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

	rate, err := uc.fx.GetRate(ctx, t.Currency, baseCurrency)
	if err != nil {
		return 0, fmt.Errorf("fx rate %s to %s: %w", t.Currency, baseCurrency, err)
	}

	return rate.Rate, nil
}

// load returns the instrument and what is held of it. An unknown instrument or position means
// nothing is held; a known instrument's stored identity replaces the one derived from the trade.
func (uc *ImportTradesUseCase) load(
	ctx context.Context,
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

	existing, err := uc.instruments.GetBySymbolAndExchange(ctx, t.Symbol, t.Exchange)
	if errors.Is(err, domain.ErrInstrumentNotFound) {
		return fresh, nil
	}
	if err != nil {
		return tracked{}, err
	}

	pos, err := uc.positions.Get(ctx, portfolioID, existing.ID)
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
		return t, fmt.Errorf(
			"could not resolve instrument metadata for %q; add instrument_type, exchange and quote_currency columns",
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

// fxRate converts the trade currency to the instrument's quote currency.
// An explicit rate on the trade wins; otherwise the current rate is used.
func (uc *ImportTradesUseCase) fxRate(ctx context.Context, t domain.TradeInput) (float64, error) {
	if t.Currency == t.QuoteCurrency {
		return 1, nil
	}
	if t.FXRate > 0 {
		return t.FXRate, nil
	}

	rate, err := uc.fx.GetRate(ctx, t.Currency, t.QuoteCurrency)
	if err != nil {
		return 0, fmt.Errorf("fx rate %s to %s: %w", t.Currency, t.QuoteCurrency, err)
	}

	return rate.Rate, nil
}

func (uc *ImportTradesUseCase) apply(ctx context.Context, portfolioID int64, p plannedTrade) error {
	instrument, err := uc.instruments.GetOrCreate(ctx, p.instrument)
	if err != nil {
		return fmt.Errorf("get or create instrument %q: %w", p.trade.Symbol, err)
	}

	if err := uc.savePosition(ctx, portfolioID, instrument, p.after); err != nil {
		return err
	}

	return uc.ledger.Create(ctx, repository.Trade{
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

func (uc *ImportTradesUseCase) savePosition(
	ctx context.Context,
	portfolioID int64,
	instrument repository.Instrument,
	after domain.TradePosition,
) error {
	if after.IsClosed() {
		if err := uc.positions.Delete(ctx, portfolioID, instrument.ID); err != nil {
			return fmt.Errorf("close position %q: %w", instrument.Symbol, err)
		}
		return nil
	}

	err := uc.positions.Create(ctx, repository.CreatePositionParams{
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
