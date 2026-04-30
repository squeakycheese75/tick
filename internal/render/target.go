package render

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderSetTarget(w io.Writer, out domain.SetTargetUseCaseOutput) error {
	_, err := fmt.Fprintf(
		w,
		"Set %s target for %s in portfolio %s: %.2f %s\n",
		out.Type,
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

	for _, t := range out.Targets {
		_, err := fmt.Fprintf(
			w,
			"%-6s %-11s %12.2f %s\n",
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
