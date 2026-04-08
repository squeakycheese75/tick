package cli

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderPortfolioSummary(w io.Writer, s domain.Summary) error {
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
