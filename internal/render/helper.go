package render

import "time"

import "math"

func isEffectivelyZero(v float64) bool {
	return math.Abs(v) < 0.005
}

func isEffectivelyZeroPct(v float64) bool {
	return math.Abs(v) < 0.00005
}

func shouldShowChange(abs, pct float64) bool {
	return !isEffectivelyZero(abs) || !isEffectivelyZeroPct(pct)
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return s[:max-1] + "…"
}

func renderKeyValue(out *writer, key, value string) {
	out.printf("%-13s %s\n", key, value)
}

func colorize(v float64, s string) string {
	switch {
	case v > 0:
		return ansiGreen + s + ansiReset
	case v < 0:
		return ansiRed + s + ansiReset
	default:
		return s
	}
}

// pricesAsOfNote says when prices are from an earlier day than now, such as before the market
// opens or at a weekend, so a last-session move is not mistaken for today's. Empty otherwise.
func pricesAsOfNote(asOf, now time.Time) string {
	if asOf.IsZero() || now.IsZero() {
		return ""
	}

	local := asOf.In(now.Location())
	if !local.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
		return ""
	}

	return "Prices as of " + local.Format("Mon 2 Jan")
}
