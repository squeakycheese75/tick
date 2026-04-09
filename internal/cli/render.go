package cli

import (
	"fmt"
	"io"

	"github.com/squeakycheese75/tick/internal/usecase"
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

func RenderGetPortfolioSummary(w io.Writer, s usecase.GetPortfolioSummaryUsecaseOutput) error {
	out := &writer{w: w}

	out.printf("Portfolio: %s\n\n", s.PortfolioName)
	out.printf("Base currency: %s\n", s.BaseCurrency)
	out.printf("Total value: %.2f\n\n", s.TotalValue)

	if len(s.Positions) == 0 {
		out.println("No positions")
		return nil
	}

	out.println("Positions:")

	for _, p := range s.Positions {
		out.printf(
			"- %s qty=%.4f price=%.2f %s fx=%.4f value=%.2f %s weight=%.2f%%\n",
			p.Ticker,
			p.Quantity,
			p.CurrentPrice,
			p.InstrumentCurrency,
			p.FXRate,
			p.MarketValueBase,
			s.BaseCurrency,
			p.Weight*100,
		)
	}

	return out.err
}

func RenderCreatePortfolio(w io.Writer, s usecase.CreatePortfolioUsecaseOutout) error {
	out := &writer{w: w}

	out.printf(
		"Portfolio %q saved (base currency: %s)\n",
		s.PortfolioName,
		s.BaseCurrency,
	)
	return out.err
}

func RenderAddPortfolioPosition(w io.Writer, s usecase.AddPositionToPortfolioUseCaseOutput) error {
	out := &writer{w: w}

	out.printf("Saved %s in portfolio %s: qty=%.4f avg_cost=%.2f %s\n", s.Ticker, s.PortfolioName, s.Qty, s.AvgCost, s.Currency)

	return out.err
}

func RenderGetPortfolioRisk(w io.Writer, s usecase.GetPortfolioRiskOutput) error {
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

func RenderGetDailyBrief(w io.Writer, s usecase.GetDailyBriefOutput) error {
	out := &writer{w: w}

	out.printf("tick daily\n\n")
	out.printf("Portfolio: %s\n", s.PortfolioName)
	out.printf("Base currency: %s\n", s.BaseCurrency)
	out.printf("Total value: %.2f\n\n", s.TotalValue)

	out.printf("Top holdings\n")
	if len(s.TopHoldings) == 0 {
		out.println("- No positions")
	} else {
		for _, h := range s.TopHoldings {
			out.printf(
				"- %s  %.2f%%  %.2f %s\n",
				h.Ticker,
				h.Weight*100,
				h.MarketValueBase,
				s.BaseCurrency,
			)
		}
	}

	out.println("\nRisk")
	if s.Risk.LargestPosition == "" {
		out.printf("- No risk data available\n")
	} else {
		out.printf("- Largest position: %s (%.2f%%)\n", s.Risk.LargestPosition, s.Risk.LargestWeight*100)
		out.printf("- Top 3 concentration: %.2f%%\n", s.Risk.Top3Concentration*100)
		for _, observation := range s.Risk.Observations {
			out.printf("- Observation: %s\n", observation)
		}
	}

	out.println("\nNews")
	if len(s.News) == 0 {
		out.printf("- No news\n")
	} else {
		for _, group := range s.News {
			if len(group.Headlines) == 0 {
				out.printf("- %s: no recent headlines\n", group.Ticker)
				continue
			}
			out.printf("- %s:\n", group.Ticker)
			for _, headline := range group.Headlines {
				out.printf("  - %s\n", headline.Title)
			}
		}
	}

	out.println("\nAttention")
	for _, item := range s.Attention {
		out.printf("- %s\n", item)
	}

	return out.err
}
