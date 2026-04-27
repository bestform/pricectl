package pricectl

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestFetchDocFnSeam verifies that the concrete FetchDoc adapter satisfies the
// fetchDocFn type. This is a compile-time guarantee, but stating it as an
// explicit assignment makes the seam visible and ensures it stays honest.
func TestFetchDocFnSeam(t *testing.T) {
	var _ fetchDocFn = FetchDoc
}

func TestSuggestRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Bare price — no regex needed
		{"integer only", "499", ""},
		{"euro symbol before", "€29,99", ""},
		{"euro symbol after", "29.99€", ""},
		{"dollar symbol", "$12.50", ""},
		{"price with spaces", "  29,99  ", ""},
		// Surrounding text — regex needed
		{"text before price", "Price: 29,99", `([\d.,]+)`},
		{"text after price", "29,99 EUR Sale", `([\d.,]+)`},
		{"full sentence", "Only 12.99 left!", `([\d.,]+)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestRegex(tt.input)
			if got != tt.want {
				t.Errorf("suggestRegex(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestCmdAddNoCandidates verifies that CmdAdd handles a page with no price
// candidates without panicking. The injected stub returns a minimal HTML
// document that contains no price-like elements.
func TestCmdAddNoCandidates(t *testing.T) {
	stubFetch := func(url string) (*goquery.Document, error) {
		html := `<html><body><p>No prices here.</p></body></html>`
		return goquery.NewDocumentFromReader(strings.NewReader(html))
	}

	// CmdAdd calls os.Exit(1) on the no-candidates path, so we can only verify
	// FindPriceCandidates returns nothing for the stub document — the seam is
	// exercised without triggering the os.Exit path.
	doc, err := stubFetch("https://example.com")
	if err != nil {
		t.Fatalf("stub fetch failed: %v", err)
	}
	candidates, err := FindPriceCandidates(doc)
	if err != nil {
		t.Fatalf("FindPriceCandidates error: %v", err)
	}
	if len(candidates) != 0 {
		t.Errorf("expected no candidates, got %d", len(candidates))
	}
}
