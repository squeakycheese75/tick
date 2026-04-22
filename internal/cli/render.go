package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
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

const (
	ansiReset = "\033[0m"
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
)

func formatChangePercent(v float64) string {
	arrow := "→"
	color := ""
	reset := ""

	switch {
	case v > 0:
		arrow = "↑"
		color = ansiGreen
		reset = ansiReset
	case v < 0:
		arrow = "↓"
		color = ansiRed
		reset = ansiReset
	}

	return fmt.Sprintf("%s%s %+.2f%%%s", color, arrow, v, reset)
}

func RenderGetPortfolioSummary(w io.Writer, s domain.GetPortfolioSummaryUsecaseOutput) error {
	out := &writer{w: w}

	out.printf("Portfolio: %s\n\n", s.PortfolioName)
	out.printf("Base currency: %s\n", s.BaseCurrency)
	out.printf("Total value: %.2f %s\n", s.TotalValue, s.BaseCurrency)
	out.printf("Total cost: %.2f %s\n", s.TotalCost, s.BaseCurrency)
	out.printf("Total PnL: %.2f %s (%.2f%%)\n\n", s.TotalPnL, s.BaseCurrency, s.TotalPnLPct*100)

	if len(s.Positions) == 0 {
		out.println("No positions")
		return out.err
	}

	out.println("Positions:")
	out.println("TICKER   QTY         PRICE         VALUE         COST          PNL           PNL %")
	out.println("------   ----------  ------------  ------------  ------------  ------------  -------")

	for _, p := range s.Positions {
		priceStr := fmt.Sprintf("%.2f %s", p.QuotedPrice, p.InstrumentCurrency)
		valueStr := fmt.Sprintf("%.2f %s", p.MarketValueBase, s.BaseCurrency)
		costStr := fmt.Sprintf("%.2f %s", p.CostBasisBase, s.BaseCurrency)
		pnlStr := fmt.Sprintf("%.2f %s", p.UnrealizedPnL, s.BaseCurrency)

		out.printf(
			"%-6s   %10.4f  %12s  %12s  %12s  %12s  %6.2f%%\n",
			p.Symbol,
			p.Quantity,
			priceStr,
			valueStr,
			costStr,
			pnlStr,
			p.UnrealizedPnLPct*100,
		)
	}

	return out.err
}

func RenderCreatePortfolio(w io.Writer, s domain.CreatePortfolioUsecaseOutout) error {
	out := &writer{w: w}

	out.printf(
		"Portfolio %q saved (base currency: %s)\n",
		s.PortfolioName,
		s.BaseCurrency,
	)
	return out.err
}

func RenderAddPortfolioPosition(w io.Writer, s domain.AddPositionToPortfolioOutput) error {
	out := &writer{w: w}

	out.printf("Saved %s in portfolio %s: qty=%.4f avg_cost=%.2f %s\n", s.Symbol, s.PortfolioName, s.Qty, s.AvgCost, s.QuoteCurrency)

	return out.err
}

func RenderGetPortfolioRisk(w io.Writer, s domain.GetPortfolioRiskOutput) error {
	out := &writer{w: w}

	out.printf("Risk summary: %s\n\n", s.PortfolioName)
	out.printf("Base currency: %s\n", s.BaseCurrency)
	out.printf("Positions: %d\n", s.PositionCount)

	if s.PositionCount > 0 {
		out.printf("Largest position: %s (%.2f%%)\n", s.LargestPosition, s.LargestWeight*100)
		out.printf("Top 3 concentration: %.2f%%\n", s.Top3Concentration*100)
	}

	if len(s.Observations) > 0 {
		out.println("\nObservations:")
		for _, observation := range s.Observations {
			out.printf("- %s\n", observation)
		}
	}

	return out.err
}

func RenderGetDailyReport(w io.Writer, s domain.GetDailyReportOutput) error {
	out := &writer{w: w}

	out.printf("Portfolio: %s\n", s.DailyReport.PortfolioName)
	// out.printf("Base currency: %s\n", s.DailyReport.BaseCurrency)
	out.printf("Total value: %.2f %s \n", s.DailyReport.TotalValue, s.DailyReport.BaseCurrency)

	if s.DailyReport.ChangeSinceLastSnapshot != nil {
		out.printf(
			"Since last snapshot: %+.2f %s (%+.2f%%)\n",
			s.DailyReport.ChangeSinceLastSnapshot.Absolute,
			s.DailyReport.BaseCurrency,
			s.DailyReport.ChangeSinceLastSnapshot.Percent*100,
		)
	}

	out.println("")

	out.printf("Top holdings\n")
	if len(s.DailyReport.TopHoldings) == 0 {
		out.println("- No positions")
	} else {
		for _, h := range s.DailyReport.TopHoldings {
			out.printf(
				"- %s  %.2f%%  %.2f %s  @ %.2f %s  %s\n",
				h.Symbol,
				h.Weight*100,
				h.MarketValueBase,
				s.DailyReport.BaseCurrency,
				h.QuotedPrice,
				h.PriceCurrency,
				formatChangePercent(h.ChangePercent),
			)

			if h.SinceLastSnapshot != nil {
				out.printf(
					"  Δ snapshot: %+.2f %s (%+.2f%%)\n",
					h.SinceLastSnapshot.ValueAbsolute,
					s.DailyReport.BaseCurrency,
					h.SinceLastSnapshot.ValuePercent*100,
				)
			}
		}
	}

	out.println("\nRisk")
	if s.DailyReport.Risk.LargestPosition == "" {
		out.printf("- No risk data available\n")
	} else {
		out.printf("- Largest position: %s (%.2f%%)\n", s.DailyReport.Risk.LargestPosition, s.DailyReport.Risk.LargestWeight*100)
		out.printf("- Top 3 concentration: %.2f%%\n", s.DailyReport.Risk.Top3Concentration*100)
		for _, observation := range s.DailyReport.Risk.Observations {
			out.printf("- Observation: %s\n", observation)
		}
	}

	out.println("\nNews")
	if len(s.DailyReport.News) == 0 {
		out.printf("- No news\n")
	} else {
		for _, group := range s.DailyReport.News {
			if len(group.Headlines) == 0 {
				continue
			}

			out.printf("- %s:\n", group.Ticker)

			for _, headline := range group.Headlines {
				out.printf("  - %s\n", headline.Title)

				if headline.URL != "" {
					// if fullLinks {
					out.printf("    🔗 %s\n", headline.URL)
					// } else {
					// out.printf("    🔗 %s\n", extractDomain(headline.URL))
					// }
				}
			}
			out.println("")
		}
	}

	out.println("\nAttention")
	for _, item := range s.DailyReport.Attention {
		out.printf("- %s\n", item)
	}

	if s.AISummary != "" {
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

func RenderTickerNews(w io.Writer, r domain.TickerNewsReport) error {
	out := &writer{w: w}

	out.printf("News for %s\n\n", r.Ticker)

	if len(r.Headlines) == 0 {
		out.println("No recent headlines")
		return out.err
	}

	for i, h := range r.Headlines {
		out.printf("- %s\n", h.Title)

		if h.URL != "" {
			out.printf("    🔗 %s\n", h.URL)
		}

		if i < len(r.Headlines)-1 {
			out.println("")
		}
	}

	return out.err
}

func renderImportPortfolio(w io.Writer, out domain.ImportPortfolioOutput) error {
	_, err := fmt.Fprintf(
		w,
		"Imported portfolio %q (%s): %d positions%s\n",
		out.PortfolioName,
		out.BaseCurrency,
		out.ImportedPositions,
		map[bool]string{true: ", portfolio created", false: ""}[out.CreatedPortfolio],
	)
	return err
}

// func extractDomain(raw string) string {
// 	u, err := url.Parse(raw)
// 	if err != nil {
// 		return raw
// 	}
// 	return u.Host
// }
