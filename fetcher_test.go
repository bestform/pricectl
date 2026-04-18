package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// docFromString parses an HTML string into a goquery document for use in tests.
func docFromString(html string) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		panic("docFromString: " + err.Error())
	}
	return doc
}

func TestParsePrice(t *testing.T) {
	valid := []struct {
		input    string
		expected int64
	}{
		{"499", 49900},
		{"29,99", 2999},
		{"29.99", 2999},
		{"1.299,00", 129900},
		{"1,299.00", 129900},
		{"9,99", 999},
		{"100", 10000},
		{"€ 49,99", 4999},
		{"$39", 3900},
		{"9.99 €", 999},
	}

	for _, c := range valid {
		got, err := parsePrice(c.input)
		if err != nil {
			t.Errorf("parsePrice(%q) unexpected error: %v", c.input, err)
			continue
		}
		if got != c.expected {
			t.Errorf("parsePrice(%q) = %d, want %d", c.input, got, c.expected)
		}
	}

	invalid := []string{
		"free",
		"abc",
		"",
		"Trial",
		"N/A",
	}

	for _, input := range invalid {
		_, err := parsePrice(input)
		if err == nil {
			t.Errorf("parsePrice(%q) expected error, got nil", input)
		}
	}
}

func TestExtractPrice(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		selector        string
		regex           string
		wantCents       int64
		wantElementHTML string
		wantErr         bool
	}{
		{
			name:            "simple match",
			html:            `<span class="price">29.99</span>`,
			selector:        ".price",
			wantCents:       2999,
			wantElementHTML: `<span class="price">29.99</span>`,
		},
		{
			name:            "first element unparseable, second is valid",
			html:            `<span class="price">Trial</span><span class="price">$39</span>`,
			selector:        ".price",
			wantCents:       3900,
			wantElementHTML: `<span class="price">$39</span>`,
		},
		{
			name:            "regex extracts number from surrounding text",
			html:            `<span class="price">Now only 14,99 EUR!</span>`,
			selector:        ".price",
			regex:           `([\d,]+)`,
			wantCents:       1499,
			wantElementHTML: `<span class="price">Now only 14,99 EUR!</span>`,
		},
		{
			name:     "selector matches nothing",
			html:     `<div>no price here</div>`,
			selector: ".price",
			wantErr:  true,
		},
		{
			name:     "selector matches element with no parseable price",
			html:     `<span class="price">free</span>`,
			selector: ".price",
			wantErr:  true,
		},
		{
			name:     "invalid regex",
			html:     `<span class="price">29.99</span>`,
			selector: ".price",
			regex:    `([invalid`,
			wantErr:  true,
		},
		{
			name:            "german price format",
			html:            `<span class="price">1.299,00</span>`,
			selector:        ".price",
			wantCents:       129900,
			wantElementHTML: `<span class="price">1.299,00</span>`,
		},
		{
			name:            "price with currency symbol",
			html:            `<span class="price">€ 49,99</span>`,
			selector:        ".price",
			wantCents:       4999,
			wantElementHTML: `<span class="price">€ 49,99</span>`,
		},
		{
			name:            "multiple elements, picks first valid",
			html:            `<span class="price">free</span><span class="price">19.99</span><span class="price">24.99</span>`,
			selector:        ".price",
			wantCents:       1999,
			wantElementHTML: `<span class="price">19.99</span>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := docFromString(tt.html)
			got, elementHTML, err := extractPrice(doc, tt.selector, tt.regex)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (price: %d)", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.wantCents {
				t.Errorf("got %d cents, want %d cents", got, tt.wantCents)
			}
			if elementHTML != tt.wantElementHTML {
				t.Errorf("got elementHTML %q, want %q", elementHTML, tt.wantElementHTML)
			}
		})
	}
}
