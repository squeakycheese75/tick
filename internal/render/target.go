package render

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

const targetListFormat = "%-4s  %-8s  %-14s  %12s  %-3s\n"
const targetRowFormat = "%-4d  %-8s  %-14s  %12.2f  %-3s\n"

func RenderSetTarget(w io.Writer, out domain.SetTargetUseCaseOutput) error {
	_, err := fmt.Fprintf(
		w,
		"Set %s target %d for %s in portfolio %s: %.2f %s\n",
		out.Type,
		out.TargetID,
		out.Symbol,
		out.PortfolioName,
		out.TargetPrice,
		out.QuoteCurrency,
	)
	return err
}
func RenderListTargets(w io.Writer, out domain.ListTargetsUseCaseOutput) error {
	if len(out.Targets) == 0 {
		_, err := fmt.Fprintf(w, "No targets set for portfolio %q\n", out.PortfolioName)
		return err
	}

	_, err := fmt.Fprintf(w, "Targets for %s\n", out.PortfolioName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, targetListFormat, "ID", "SYMBOL", "TYPE", "PRICE", "CCY")
	if err != nil {
		return err
	}

	for _, t := range out.Targets {
		_, err := fmt.Fprintf(
			w,
			targetRowFormat,
			t.ID,
			t.Symbol,
			t.Type,
			t.TargetPrice,
			t.QuoteCurrency,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
