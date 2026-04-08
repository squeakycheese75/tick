package cli

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/cmd/usecase"
)

type writer struct {
	w   io.Writer
	err error
}

func (w *writer) printf(format string, args ...any) {
	if w.err != nil {
		return
	}
	_, w.err = fmt.Fprintf(w.w, format, args...)
}

func (w *writer) println(s string) {
	w.printf("%s\n", s)
}

func RenderGetPortfolioSummary(w io.Writer, s usecase.GetPortfolioSummaryUsecaseOutput) error {
	out := &writer{w: w}

	out.printf("Portfolio: %s\n\n", s.PortfolioName)
	out.printf("Base currency: %s\n", s.BaseCurrency)
	out.printf("Total value: %.2f\n\n", s.TotalValue)

	if len(s.Positions) == 0 {
		out.println("No positions")
		return nil
	}

	out.println("Positions:")

	for _, p := range s.Positions {
		out.printf(
			"- %s qty=%.4f price=%.2f %s fx=%.4f value=%.2f %s weight=%.2f%%\n",
			p.Ticker,
			p.Quantity,
			p.CurrentPrice,
			p.InstrumentCurrency,
			p.FXRate,
			p.MarketValueBase,
			s.BaseCurrency,
			p.Weight*100,
		)
	}

	return out.err
}

func RenderCreatePortfolio(w io.Writer, s usecase.CreatePortfolioUsecaseOutout) error {
	out := &writer{w: w}

	out.printf(
		"Portfolio %q saved (base currency: %s)\n",
		s.Name,
		s.BaseCurrency,
	)
	return out.err
}

func RenderAddPortfolioPosition(w io.Writer, s usecase.AddPositionToPortfolioUseCaseOutput) error {
	out := &writer{w: w}

	out.printf("Saved %s in portfolio %s: qty=%.4f avg_cost=%.2f %s\n", s.Ticker, s.PortfolioName, s.Qty, s.AvgCost, s.Currency)

	return out.err
}
