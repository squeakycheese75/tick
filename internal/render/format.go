package render

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	ansiReset = "\033[0m"
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
)

var printer = message.NewPrinter(language.English)

func formatAmount(v float64) string {
	return printer.Sprintf("%.2f", v)
}

func formatSignedAmount(v float64) string {
	return printer.Sprintf("%+.2f", v)
}

func formatMoney(v float64, ccy string) string {
	return fmt.Sprintf("%s %s", formatAmount(v), ccy)
}

func formatSignedMoney(v float64, ccy string) string {
	return fmt.Sprintf("%s %s", formatSignedAmount(v), ccy)
}

func formatQuantity(q float64) string {
	return formatQuantityDP(q, quantityDecimals([]float64{q}))
}

func formatQuantityDP(q float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, q)
}

// quantityDecimals picks one number of decimals for a whole column, enough for its smallest
// quantity, so the decimal points line up when the column is right-aligned.
func quantityDecimals(quantities []float64) int {
	decimals := 0

	for _, q := range quantities {
		switch {
		case q < 1 && q != 0:
			decimals = max(decimals, 6)
		case q < 100:
			decimals = max(decimals, 4)
		}
	}

	return decimals
}

// func formatPercentFromRatio(v float64) string {
// 	return printer.Sprintf("%.2f%%", v*100)
// }

func formatSignedPercentFromRatio(v float64) string {
	return printer.Sprintf("%+.2f%%", v*100)
}

// func formatPercent(v float64) string {
// 	return printer.Sprintf("%.2f%%", v)
// }

func formatSignedPercent(v float64) string {
	return printer.Sprintf("%+.2f%%", v)
}

func formatSignedMoneyColored(v float64, ccy string, color bool) string {
	base := formatSignedMoney(v, ccy)
	if !color {
		return base
	}
	return colorize(v, base)
}

func formatSignedPercentColored(v float64, color bool) string {
	base := formatSignedPercent(v)
	if !color {
		return base
	}
	return colorize(v, base)
}

func formatChangePercent(v float64, color bool) string {
	arrow := "→"
	col := ""
	reset := ""

	if color {
		switch {
		case v > 0:
			arrow = "↑"
			col = ansiGreen
			reset = ansiReset
		case v < 0:
			arrow = "↓"
			col = ansiRed
			reset = ansiReset
		}
	}

	return fmt.Sprintf("%s%s %+.2f%%%s", col, arrow, v, reset)
}
