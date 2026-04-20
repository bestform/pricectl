package pricectl

import (
	"strings"
	"testing"
)

func TestFormatCents(t *testing.T) {
	cases := []struct {
		cents    int64
		expected string
	}{
		{0, "0.00"},
		{1, "0.01"},
		{99, "0.99"},
		{100, "1.00"},
		{2999, "29.99"},
		{49900, "499.00"},
		{129900, "1299.00"},
	}

	for _, c := range cases {
		got := formatCents(c.cents)
		if got != c.expected {
			t.Errorf("formatCents(%d) = %q, want %q", c.cents, got, c.expected)
		}
	}
}

func TestFormatDiff(t *testing.T) {
	cases := []struct {
		prev     int64
		curr     int64
		contains string // expected substring in output
	}{
		{4999, 2999, "-20.00"},
		{2999, 4999, "+20.00"},
		{2999, 2999, "+0.00"},
	}

	for _, c := range cases {
		got := formatDiff(c.prev, c.curr)
		// Strip ANSI codes for comparison
		plain := stripANSI(got)
		if !strings.Contains(plain, c.contains) {
			t.Errorf("formatDiff(%d, %d) = %q, want it to contain %q", c.prev, c.curr, plain, c.contains)
		}
	}
}

func TestPriceArrow(t *testing.T) {
	cases := []struct {
		prev     int64
		curr     int64
		contains string
	}{
		{4999, 2999, "▼"},
		{2999, 4999, "▲"},
		{2999, 2999, "="},
	}

	for _, c := range cases {
		got := priceArrow(c.prev, c.curr)
		plain := stripANSI(got)
		if !strings.Contains(plain, c.contains) {
			t.Errorf("priceArrow(%d, %d) = %q, want it to contain %q", c.prev, c.curr, plain, c.contains)
		}
	}
}

// stripANSI removes ANSI escape codes from a string for plain-text comparison.
func stripANSI(s string) string {
	result := strings.Builder{}
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm'
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}
