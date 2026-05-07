package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

func DailyReport(w io.Writer, s domain.GetDailyReportOutput, opts DailyReportOptions) error {
	out := &writer{w: w}
	r := s.DailyReport

	renderPortfolioSummary(out, r.Portfolio, opts.Summary)
	renderHoldingSummary(out, r.TopHoldings, r.Portfolio.BaseCurrency, opts.Holdings)
	renderValuationIssuesSummary(out, r.ValuationIssues)
	renderRiskSummary(out, r.Risk, opts.Risk)
	renderNewsSummary(out, r.News, opts.News)
	renderTargets(out, r.Targets)
	renderAttentionSummary(out, r.Attention)
	renderAISummary(out, s.AISummary, opts.AI)

	return out.err
}

func renderPortfolioSummary(
	out *writer,
	r domain.PortfolioSummary,
	opts SummaryOptions,
) {
	out.printf("%s  %s", r.Name, formatMoney(r.TotalValue, r.BaseCurrency))

	if opts.ShowSnapshotDelta &&
		r.Change != nil &&
		(!opts.HideZeroDelta || shouldShowChange(
			r.Change.Absolute,
			r.Change.Percent,
		)) {

		// Likely first snapshot / no baseline.
		if r.Change.Absolute != 0 && r.Change.Percent == 0 {
			out.printf(
				"  new snapshot (%s)",
				formatSignedMoney(r.Change.Absolute, r.BaseCurrency),
			)
		} else {
			out.printf(
				"  Δ %s (%s)",
				formatSignedMoney(r.Change.Absolute, r.BaseCurrency),
				formatSignedPercentFromRatio(r.Change.Percent),
			)
		}
	}

	out.println("")
}

func renderHoldingSummary(
	out *writer,
	r domain.HoldingSummary,
	baseCurrency string,
	opts HoldingsOptions,
) {
	renderHoldingRows(out, r, baseCurrency, opts, false)
}

func renderMoversSummary(
	out *writer,
	r domain.HoldingSummary,
	baseCurrency string,
	opts HoldingsOptions,
) {
	renderHoldingRows(out, r, baseCurrency, opts, true)
}

func renderHoldingRows(
	out *writer,
	r domain.HoldingSummary,
	baseCurrency string,
	opts HoldingsOptions,
	showAbsChange bool,
) {
	out.println("")

	title := opts.Title
	if title == "" {
		if opts.ShowTop > 0 {
			title = "Top Holdings"
		} else {
			title = "Holdings"
		}
	}

	out.println(title)
	if len(r.Holdings) == 0 {
		out.println("No priced positions")
		return
	}

	for _, h := range r.Holdings {
		absChange := ""
		if showAbsChange {
			absChange = fmt.Sprintf(
				"  %s",
				formatSignedMoneyColored(h.ChangeAbsolute, baseCurrency, opts.Color),
			)
		}

		snapshot := ""
		if opts.ShowSnapshotDelta &&
			h.SinceLastSnapshot != nil &&
			(!opts.HideZeroDelta || shouldShowChange(
				h.SinceLastSnapshot.Absolute,
				h.SinceLastSnapshot.Percent,
			)) {
			snapshot = fmt.Sprintf(
				"  Δsnap %s (%s)",
				formatSignedMoneyColored(h.SinceLastSnapshot.Absolute, baseCurrency, opts.Color),
				formatSignedPercentColored(h.SinceLastSnapshot.Percent, opts.Color),
			)
		}

		out.printf(
			"%-6s %10s %7.2f%% %16s  @ %13s  %s%s%s\n",
			h.Symbol,
			formatQuantity(h.Quantity),
			h.Weight*100,
			formatMoney(h.MarketValueBase, baseCurrency),
			formatMoney(h.QuotedPrice, h.PriceCurrency),
			formatChangePercent(h.ChangePercent, opts.Color),
			absChange,
			snapshot,
		)
	}
}

func renderValuationIssuesSummary(out *writer, issues []domain.ValuationIssue) {
	if len(issues) == 0 {
		return
	}

	out.println("")
	out.println("Unpriced")

	for _, issue := range issues {
		out.printf(
			"%-28s %12s %-8s %s\n",
			issue.Symbol,
			formatQuantity(issue.Quantity),
			issue.InstrumentType,
			issue.Message,
		)
	}
}

func renderRiskSummary(out *writer, r domain.RiskSummary, opts RiskOptions) {
	out.println("")

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

func renderNewsSummary(out *writer, groups []domain.NewsSummary, opts NewsOptions) {
	out.println("")
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

func renderTargets(out *writer, targets []domain.TargetStatus) {
	if len(targets) == 0 {
		return
	}

	out.println("")
	out.println("Targets")

	for _, t := range targets {
		marker := ""
		if t.Hit {
			marker = "  ! hit"
		}

		out.printf(
			"%-6s %-11s target %12.2f %s  current %12.2f %s%s\n",
			t.Symbol,
			t.Type,
			t.TargetPrice,
			t.Currency,
			t.CurrentPrice,
			t.Currency,
			marker,
		)
	}
}

func renderAttentionSummary(out *writer, attention []string) {
	if len(attention) == 0 {
		return
	}

	out.println("")
	out.println("Attention")

	for _, item := range attention {
		out.printf("- %s\n", item)
	}
}

func renderAISummary(out *writer, summary string, opts AIOptions) {
	if !opts.Show || strings.TrimSpace(summary) == "" {
		return
	}

	out.println("")
	out.println("AI Summary")

	for _, line := range strings.Split(summary, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		out.println(line)
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
