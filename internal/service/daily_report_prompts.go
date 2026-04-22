package service

import (
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

func buildDailyReportSystemPrompt() string {
	return strings.TrimSpace(`
You are a cautious portfolio analyst.

Your job is to summarize the supplied portfolio facts clearly and briefly.
Only use the information provided.
Do not invent any figures, positions, news, or market context.
Do not provide buy or sell recommendations.
Focus on concentration, notable daily moves, and what deserves attention today.
Return 3 to 5 concise bullet points.
`)
}

func buildDailyReportUserPrompt(dailyReport domain.DailyReport) string {
	var b strings.Builder

	b.WriteString("Portfolio daily brief\n\n")

	b.WriteString(fmt.Sprintf("Portfolio: %s\n", dailyReport.PortfolioName))
	b.WriteString(fmt.Sprintf("Base currency: %s\n", dailyReport.BaseCurrency))
	b.WriteString(fmt.Sprintf("Total value: %.2f %s\n\n", dailyReport.TotalValue, dailyReport.BaseCurrency))

	b.WriteString("Top holdings:\n")
	if len(dailyReport.TopHoldings) == 0 {
		b.WriteString("- No positions\n")
	} else {
		for _, h := range dailyReport.TopHoldings {
			b.WriteString(fmt.Sprintf(
				"- %s: weight %.2f%%, value %.2f %s, quoted price %.2f %s, daily move %+.2f%%\n",
				h.Symbol,
				h.Weight*100,
				h.MarketValueBase,
				dailyReport.BaseCurrency,
				h.QuotedPrice,
				h.PriceCurrency,
				h.ChangePercent,
			))
		}
	}
	b.WriteString("\n")

	b.WriteString("Risk:\n")
	if dailyReport.Risk.LargestPosition == "" {
		b.WriteString("- No risk data available\n")
	} else {
		b.WriteString(fmt.Sprintf("- Largest position: %s (%.2f%%)\n", dailyReport.Risk.LargestPosition, dailyReport.Risk.LargestWeight*100))
		b.WriteString(fmt.Sprintf("- Top 3 concentration: %.2f%%\n", dailyReport.Risk.Top3Concentration*100))
		for _, obs := range dailyReport.Risk.Observations {
			b.WriteString(fmt.Sprintf("- %s\n", obs))
		}
	}
	b.WriteString("\n")

	b.WriteString("News:\n")
	if len(dailyReport.News) == 0 {
		b.WriteString("- No news available\n")
	} else {
		for _, group := range dailyReport.News {
			if len(group.Headlines) == 0 {
				b.WriteString(fmt.Sprintf("- %s: no recent headlines\n", group.Ticker))
				continue
			}
			b.WriteString(fmt.Sprintf("- %s:\n", group.Ticker))
			for _, headline := range group.Headlines {
				b.WriteString(fmt.Sprintf("  - %s\n", headline.Title))
			}
		}
	}
	b.WriteString("\n")

	b.WriteString("Attention:\n")
	if len(dailyReport.Attention) == 0 {
		b.WriteString("- None\n")
	} else {
		for _, item := range dailyReport.Attention {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	b.WriteString("\n")

	b.WriteString("Task:\n")
	b.WriteString("Summarize what matters today in 3 to 5 concise bullet points.\n")

	return b.String()
}
