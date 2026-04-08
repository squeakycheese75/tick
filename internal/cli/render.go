package cli

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/cmd/usecase"
)

func RenderGetPortfolioSummary(w io.Writer, s usecase.GetPortfolioSummaryUsecaseOutput) error {
	fmt.Fprintf(w, "Portfolio: %s\n\n", s.PortfolioName)
	fmt.Fprintf(w, "Base currency: %s\n", s.BaseCurrency)
	fmt.Fprintf(w, "Total value: %.2f\n\n", s.TotalValue)

	if len(s.Positions) == 0 {
		fmt.Fprintln(w, "No positions")
		return nil
	}

	fmt.Fprintln(w, "Positions:")

	for _, p := range s.Positions {
		fmt.Fprintf(
			w,
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

	return nil
}

func RenderCreatePortfolio(w io.Writer, s usecase.CreatePortfolioUsecaseOutout) error {
	fmt.Fprintf(
		w,
		"Portfolio %q saved (base currency: %s)\n",
		s.Name,
		s.BaseCurrency,
	)
	return nil
}

func RenderAddPortfolioPosition(w io.Writer, s usecase.AddPositionToPortfolioUseCaseOutput) error {
	fmt.Fprintf(w, "Saved %s in portfolio %s: qty=%.4f avg_cost=%.2f %s\n", s.Ticker, s.PortfolioName, s.Qty, s.AvgCost, s.Currency)

	return nil
}
