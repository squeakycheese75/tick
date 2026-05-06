package render

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/internal/domain"
)

func RenderPortfolioSummary(w io.Writer, s domain.PortfoloSummaryReport, opts PortfolioSummaryOptions) error {
	out := &writer{w: w}

	if opts.ShowHeader {
		renderPortfolioTitle(out, s.PortfolioName)
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

func renderPortfolioTitle(out *writer, portfolioName string) {
	out.println(portfolioName)
}

func renderPortfolioTotals(out *writer, s domain.PortfoloSummaryReport, opts PortfolioSummaryOptions) {
	renderKeyValue(out, "Base currency", s.BaseCurrency)
	renderKeyValue(out, "Total value", formatMoney(s.TotalValue, s.BaseCurrency))
	renderKeyValue(out, "Total cost", formatMoney(s.TotalCost, s.BaseCurrency))

	totalPnL := formatSignedMoneyColored(s.TotalPnL, s.BaseCurrency, opts.Color)
	totalPnLPct := formatSignedPercentColored(s.TotalPnLPct*100, opts.Color)
	renderKeyValue(out, "Total PnL", totalPnL+"  ("+totalPnLPct+")")
}

func renderPortfolioPositions(out *writer, s domain.PortfoloSummaryReport, opts PortfolioSummaryOptions) {
	out.println("Positions")

	if len(s.Positions) == 0 {
		out.println("No positions")
		return
	}

	out.printf(
		"%-6s %10s  %16s  %16s  %16s  %17s  %8s\n",
		"TICKER",
		"QTY",
		"PRICE",
		"VALUE",
		"COST",
		"PNL",
		"PNL %",
	)

	for _, p := range s.Positions {
		priceStr := fmt.Sprintf("%16s", formatMoney(p.QuotedPrice, p.InstrumentCurrency))
		valueStr := fmt.Sprintf("%16s", formatMoney(p.MarketValueBase, s.BaseCurrency))
		costStr := fmt.Sprintf("%16s", formatMoney(p.CostBasisBase, s.BaseCurrency))

		pnlStr := fmt.Sprintf("%17s", formatSignedMoney(p.UnrealizedPnL, s.BaseCurrency))
		if opts.Color {
			pnlStr = colorize(p.UnrealizedPnL, pnlStr)
		}

		pnlPctStr := fmt.Sprintf("%8s", formatSignedPercent(p.UnrealizedPnLPct*100))
		if opts.Color {
			pnlPctStr = colorize(p.UnrealizedPnLPct, pnlPctStr)
		}

		out.printf(
			"%-6s %10.4f  %s  %s  %s  %s  %s\n",
			p.Symbol,
			p.Quantity,
			priceStr,
			valueStr,
			costStr,
			pnlStr,
			pnlPctStr,
		)
	}
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
