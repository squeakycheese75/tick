package render

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderSetTarget(w io.Writer, out domain.SetTargetUsecaseOutput) error {
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
