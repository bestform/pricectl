package main

import "fmt"

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
)

func green(s string) string  { return colorGreen + s + colorReset }
func red(s string) string    { return colorRed + s + colorReset }
func yellow(s string) string { return colorYellow + s + colorReset }
func bold(s string) string   { return colorBold + s + colorReset }

// formatCents converts a cent value to a human-readable price string.
// e.g. 2999 -> "29.99"
func formatCents(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// priceArrow returns a colored arrow indicating price direction.
func priceArrow(prev, curr int64) string {
	switch {
	case curr < prev:
		return green("▼")
	case curr > prev:
		return red("▲")
	default:
		return "="
	}
}

// formatDiff returns a colored string showing the price difference.
func formatDiff(prev, curr int64) string {
	diff := curr - prev
	s := fmt.Sprintf("%+.2f", float64(diff)/100)
	if diff < 0 {
		return green(s)
	} else if diff > 0 {
		return red(s)
	}
	return s
}
