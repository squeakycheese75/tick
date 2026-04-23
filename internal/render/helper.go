package render

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
