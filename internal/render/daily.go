package render

import (
	"io"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

func DailyReport(w io.Writer, s domain.GetDailyReportOutput, opts DailyReportOptions) error {
	out := &writer{w: w}
	r := s.DailyReport

	renderPortfolioSummary(out, r.Portfolio, opts.Summary)
	out.println("")

	renderHoldingSummary(out, r.TopHoldings, r.Portfolio.BaseCurrency, opts.Holdings)
	out.println("")

	renderRiskSummary(out, r.Risk, opts.Risk)
	out.println("")

	renderNewsSummary(out, r.News, opts.News)

	if opts.ShowAttention && len(r.Attention) > 0 {
		out.println("")
		out.println("Attention")
		for _, item := range r.Attention {
			out.printf("- %s\n", item)
		}
	}

	if opts.AI.Show && s.AISummary != "" {
		out.println("")
		out.println("AI Summary")
		for _, line := range strings.Split(s.AISummary, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			out.println(line)
		}
	}

	return out.err
}

func renderPortfolioSummary(out *writer, r domain.PortfolioSummary, opts SummaryOptions) {
	out.printf("%s  %s", r.Name, formatMoney(r.TotalValue, r.BaseCurrency))

	if opts.ShowSnapshotDelta &&
		r.Change != nil &&
		(!opts.HideZeroDelta || shouldShowChange(
			r.Change.Absolute,
			r.Change.Percent,
		)) {
		out.printf(
			"  Δ %s (%s)",
			formatSignedMoney(r.Change.Absolute, r.BaseCurrency),
			formatSignedPercentFromRatio(r.Change.Percent),
		)
	}
}

func renderHoldingSummary(
	out *writer,
	r domain.HoldingSummary,
	baseCurrency string,
	opts HoldingsOptions,
) {
	out.println("Holdings")
	if len(r.Holdings) == 0 {
		out.println("No positions")
		return
	}

	for _, h := range r.Holdings {
		out.printf(
			"%-5s %7.2f%% %16s  @ %16s  %s",
			h.Symbol,
			h.Weight*100,
			formatMoney(h.MarketValueBase, baseCurrency),
			formatMoney(h.QuotedPrice, h.PriceCurrency),
			formatChangePercent(h.ChangePercent, opts.Color),
		)

		if opts.ShowSnapshotDelta &&
			h.SinceLastSnapshot != nil &&
			(!opts.HideZeroDelta || shouldShowChange(
				h.SinceLastSnapshot.Absolute,
				h.SinceLastSnapshot.Percent,
			)) {
			out.printf(
				"  Δsnap %s (%s)",
				formatSignedMoneyColored(
					h.SinceLastSnapshot.Absolute,
					baseCurrency,
					opts.Color,
				),
				formatSignedPercentColored(h.SinceLastSnapshot.Percent, opts.Color),
			)
		}

		out.println("")
	}
}

func renderRiskSummary(out *writer, r domain.RiskSummary, opts RiskOptions) {
	if r.LargestPosition == "" {
		out.println("Risk   No data")
		return
	}

	if opts.Compact {
		out.printf(
			"Risk   Largest: %s (%.2f%%)   Top 3: %.2f%%",
			r.LargestPosition,
			r.LargestWeight*100,
			r.Top3Concentration*100,
		)

		if label := riskLabel(r); label != "" {
			out.printf("   ! %s", label)
		}
		out.println("")
		return
	}

	out.println("Risk")
	out.printf("Largest: %s (%.2f%%)\n", r.LargestPosition, r.LargestWeight*100)
	out.printf("Top 3: %.2f%%\n", r.Top3Concentration*100)

	if opts.ShowObservations {
		for _, obs := range r.Observations {
			out.printf("- %s\n", obs)
		}
	}
}

func riskLabel(r domain.RiskSummary) string {
	switch {
	case r.LargestWeight >= 0.80:
		return "High concentration"
	case r.Top3Concentration >= 0.80:
		return "Concentrated portfolio"
	default:
		return ""
	}
}

func renderNewsSummary(out *writer, groups []domain.NewsSummary, opts NewsOptions) {
	out.println("News")

	if len(groups) == 0 {
		out.println("No news")
		return
	}

	any := false

	for _, group := range groups {
		if len(group.Headlines) == 0 {
			continue
		}

		any = true

		limit := opts.MaxHeadlines
		if limit <= 0 || limit > len(group.Headlines) {
			limit = len(group.Headlines)
		}

		for i := 0; i < limit; i++ {
			h := group.Headlines[i]
			title := h.Title
			if !opts.TruncateTitles {
				title = truncate(title, opts.HeadlineMaxLen)
			}

			if i == 0 {
				out.printf("%-5s %s\n", group.Ticker+":", title)
			} else {
				out.printf("      %s\n", title)
			}

			if opts.ShowLinks && h.URL != "" {
				out.printf("      🔗 %s\n", h.URL)
			}
		}
	}

	if !any {
		out.println("No news")
	}
}

func RenderNewsItem(w io.Writer, r domain.NewsSummary, opts NewsOptions) error {
	out := &writer{w: w}

	out.printf("News for %s\n\n", r.Ticker)

	if len(r.Headlines) == 0 {
		out.println("No recent headlines")
		return out.err
	}

	limit := opts.MaxHeadlines
	if limit <= 0 || limit > len(r.Headlines) {
		limit = len(r.Headlines)
	}

	for i := 0; i < limit; i++ {
		h := r.Headlines[i]
		title := h.Title
		if opts.TruncateTitles {
			title = truncate(title, opts.HeadlineMaxLen)
		}

		out.printf("- %s\n", title)
		if opts.ShowLinks && h.URL != "" {
			out.printf("  🔗 %s\n", h.URL)
		}
		if i < limit-1 {
			out.println("")
		}
	}

	return out.err
}
