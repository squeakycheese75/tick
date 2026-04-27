package render

import (
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderBriefReport(w io.Writer, r domain.BriefReport, opts BriefReportOptions) error {
	out := &writer{w: w}

	renderPortfolioSummary(out, r.Portfolio, opts.Summary)
	out.println("")

	renderMoversSummary(out, r.Movers, r.Portfolio.BaseCurrency, opts.Holdings)
	out.println("")

	renderNewsSummary(out, r.News, opts.News)

	return out.err
}
