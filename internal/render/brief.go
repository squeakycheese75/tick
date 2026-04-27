package render

import (
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderBriefReport(w io.Writer, r domain.BriefReport) error {
	out := &writer{w: w}

	renderPortfolioSummary(out, r.Portfolio, SummaryOptions{})
	out.println("")
	renderHoldingSummary(out, r.Movers, r.Portfolio.BaseCurrency, HoldingsOptions{})
	out.println("")

	renderNewsSummary(out, r.News, NewsOptions{})

	return out.err
}
