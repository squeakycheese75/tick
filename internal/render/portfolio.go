package render

import (
	"fmt"
	"io"
	"math"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderPortfolioSummary(w io.Writer, s domain.PortfoloSummaryReport, opts PortfolioSummaryOptions) error {
	out := &writer{w: w}

	if opts.ShowHeader {
		renderPortfolioTitle(out, s)

		if note := pricesAsOfNote(s.PricesAsOf, opts.Now); note != "" {
			out.println(note)
		}
	}

	if opts.ShowTotals {
		if opts.ShowHeader {
			out.println("")
		}
		renderPortfolioTotals(out, s, opts)
	}

	if opts.ShowPositions {
		out.println("")
		renderPortfolioPositions(out, s, opts)
	}

	return out.err
}

func renderPortfolioTitle(out *writer, s domain.PortfoloSummaryReport) {
	out.println(s.PortfolioName + " · " + s.BaseCurrency)
}

// totalsLine is one row of the totals block. Signed amounts are coloured; pct is optional.
type totalsLine struct {
	label  string
	amount float64
	signed bool
	pct    *float64
}

func renderPortfolioTotals(out *writer, s domain.PortfoloSummaryReport, opts PortfolioSummaryOptions) {
	lines := []totalsLine{
		{label: "Total value", amount: s.TotalValue},
		{label: "Total cost", amount: s.TotalCost},
		{label: "Total PnL", amount: s.TotalPnL, signed: true, pct: &s.TotalPnLPct},
	}

	tracked := countFXTracked(s)
	if tracked > 0 {
		lines = append(lines,
			totalsLine{label: "  from assets", amount: s.TotalAssetPnL, signed: true},
			totalsLine{label: "  from FX", amount: s.TotalFXPnL, signed: true},
		)
	}

	renderTotalsLines(out, lines, s.BaseCurrency, opts.Color)

	if tracked > 0 && tracked < len(s.Positions) {
		out.printf("  (FX split covers %d of %d positions; the rest have no recorded purchase rate)\n", tracked, len(s.Positions))
	}
}

// renderTotalsLines right-aligns the amounts so their decimal points line up,
// with the percentage in a column of its own.
func renderTotalsLines(out *writer, lines []totalsLine, ccy string, color bool) {
	amounts := make([]string, len(lines))
	width := 0

	for i, l := range lines {
		amounts[i] = formatMoney(l.amount, ccy)
		if l.signed {
			amounts[i] = formatSignedMoney(l.amount, ccy)
		}
		width = max(width, len(amounts[i]))
	}

	for i, l := range lines {
		amount := fmt.Sprintf("%*s", width, amounts[i])
		if l.signed && color {
			amount = colorize(l.amount, amount)
		}

		out.printf("%-13s %s%s\n", l.label, amount, totalsPct(l.pct, color))
	}
}

func totalsPct(pct *float64, color bool) string {
	if pct == nil {
		return ""
	}

	return "  " + formatSignedPercentColored(*pct*100, color)
}

func countFXTracked(s domain.PortfoloSummaryReport) int {
	tracked := 0
	for _, p := range s.Positions {
		if p.FXKnown {
			tracked++
		}
	}

	return tracked
}

// fxDust is below what two decimal places can show, so such an FX figure is treated as nothing.
const fxDust = 0.005

func renderPortfolioPositions(out *writer, s domain.PortfoloSummaryReport, opts PortfolioSummaryOptions) {
	out.println("Positions")

	if len(s.Positions) == 0 {
		out.println("No positions")
		return
	}

	out.println(positionsHeader(opts))

	for _, p := range s.Positions {
		out.println(positionRow(p, opts))
	}
}

func positionsHeader(opts PortfolioSummaryOptions) string {
	price := ""
	if opts.ShowPrices {
		price = fmt.Sprintf("  %14s", "PRICE")
	}

	return fmt.Sprintf(
		"%-6s %10s%s  %11s  %11s  %10s  %10s  %10s  %8s",
		"TICKER", "QTY", price, "VALUE", "COST", "PNL", "ASSET", "FX", "PNL %",
	)
}

func positionRow(p domain.SummaryPosition, opts PortfolioSummaryOptions) string {
	price := ""
	if opts.ShowPrices {
		price = fmt.Sprintf("  %14s", formatMoney(p.QuotedPrice, p.InstrumentCurrency))
	}

	return fmt.Sprintf(
		"%-6s %10.4f%s  %11s  %11s  %s  %s  %s  %s",
		p.Symbol,
		p.Quantity,
		price,
		formatAmount(p.MarketValueBase),
		formatAmount(p.CostBasisBase),
		pnlCell(p.UnrealizedPnL, true, opts.Color),
		pnlCell(p.AssetPnL, p.FXKnown, opts.Color),
		fxCell(p, opts.Color),
		pnlPctCell(p.UnrealizedPnLPct, opts.Color),
	)
}

// pnlCell is a signed amount, or "-" when it cannot be known.
func pnlCell(v float64, known, color bool) string {
	if !known {
		return fmt.Sprintf("%10s", "-")
	}

	cell := fmt.Sprintf("%10s", formatSignedAmount(v))
	if color {
		return colorize(v, cell)
	}

	return cell
}

// fxCell shows "-" when the purchase rate is unknown and "·" when FX has made no visible difference.
func fxCell(p domain.SummaryPosition, color bool) string {
	if p.FXKnown && math.Abs(p.FXPnL) < fxDust {
		return fmt.Sprintf("%10s", "·")
	}

	return pnlCell(p.FXPnL, p.FXKnown, color)
}

func pnlPctCell(v float64, color bool) string {
	cell := fmt.Sprintf("%8s", formatSignedPercent(v*100))
	if color {
		return colorize(v, cell)
	}

	return cell
}

func CreatePortfolio(w io.Writer, name, ccy string) error {
	out := &writer{w: w}

	out.printf("%s created (base %s)\n", name, ccy)
	return out.err
}

func PortfolioRisk(w io.Writer, s domain.GetPortfolioRiskOutput, opts PortfolioRiskOptions) error {
	out := &writer{w: w}

	out.println(s.PortfolioName)
	out.println("")

	renderKeyValue(out, "Base currency", s.BaseCurrency)
	renderKeyValue(out, "Positions", fmt.Sprintf("%d", s.PositionCount))

	if s.PositionCount > 0 {
		renderKeyValue(
			out,
			"Largest position",
			fmt.Sprintf("%s (%.2f%%)", s.LargestPosition, s.LargestWeight*100),
		)
		renderKeyValue(
			out,
			"Top 3 concentration",
			fmt.Sprintf("%.2f%%", s.Top3Concentration*100),
		)
	}

	if opts.ShowObservations && len(s.Observations) > 0 {
		out.println("")
		out.println("Observations")
		for _, observation := range s.Observations {
			out.printf("- %s\n", observation)
		}
	}

	return out.err
}

func AddPortfolioPosition(w io.Writer, s domain.Position) error {
	out := &writer{w: w}

	out.printf(
		"%s added to %s: qty=%.4f avg=%.2f %s\n",
		s.Instrument.Symbol,
		s.PortfolioName,
		s.Quantity,
		s.AvgCost,
		s.Instrument.QuoteCurrency,
	)

	return out.err
}

func ImportPortfolio(w io.Writer, out domain.ImportPortfolioOutput) error {
	writer := &writer{w: w}

	writer.printf(
		"%s imported (%s)  %d positions",
		out.PortfolioName,
		out.BaseCurrency,
		out.ImportedPositions,
	)

	if out.CreatedPortfolio {
		writer.printf("  (created)")
	}

	writer.println("")

	return writer.err
}

func ImportTrades(w io.Writer, out domain.ImportTradesOutput) error {
	writer := &writer{w: w}

	writer.printf("%s  %d trades applied", out.PortfolioName, out.Applied)

	if out.Skipped > 0 {
		writer.printf("  (%d already recorded, skipped)", out.Skipped)
	}

	writer.println("")

	return writer.err
}

func ConsumedPrices(w io.Writer, out domain.ConsumePriceUsecaseOutput) error {
	writer := &writer{w: w}

	writer.printf(
		"Consumed %s %.4f %s (%s)\n",
		out.Symbol,
		out.Price,
		out.Currency,
		out.AsOf.Format("2006-01-02"),
	)

	writer.println("")

	return writer.err
}

// PortfolioList prints each portfolio, marking the default one.
func PortfolioList(w io.Writer, out domain.ListPortfoliosOutput, defaultName string) error {
	writer := &writer{w: w}

	if len(out.Portfolios) == 0 {
		writer.println("No portfolios. Create one with: tick portfolio create <name> --base EUR")
		return writer.err
	}

	writer.printf("  %-20s %-5s %9s\n", "NAME", "BASE", "POSITIONS")

	for _, p := range out.Portfolios {
		marker := " "
		if p.Name == defaultName {
			marker = "*"
		}

		writer.printf("%s %-20s %-5s %9d\n", marker, p.Name, p.BaseCurrency, p.PositionCount)
	}

	return writer.err
}

// DefaultPortfolioSet confirms the new default portfolio.
func DefaultPortfolioSet(w io.Writer, name string) error {
	writer := &writer{w: w}
	writer.printf("Default portfolio: %s\n", name)

	return writer.err
}
