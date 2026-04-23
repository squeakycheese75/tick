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
