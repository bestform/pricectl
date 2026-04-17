package main

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Candidate is a potential price element found on a page.
type Candidate struct {
	Selector string // CSS selector to reach this element
	Text     string // raw text content of the element
	Cents    int64  // parsed price in cents (0 if unparseable)
}

// priceKeywords are substrings looked for in class/id names to find price elements.
var priceKeywords = []string{
	"price", "preis", "cost", "amount", "value",
	"offer", "angebot", "sale", "deal", "tarif",
}

// priceTextPattern matches strings that look like a price value.
var priceTextPattern = regexp.MustCompile(
	`^\s*[€$£¥]?\s*\d{1,6}([.,]\d{3})*([.,]\d{2})?\s*[€$£¥]?\s*$`,
)

// FindPriceCandidates inspects a parsed document and returns a deduplicated
// list of elements that are likely to contain a price.
func FindPriceCandidates(doc *goquery.Document) ([]Candidate, error) {
	seen := make(map[string]bool)
	var candidates []Candidate

	add := func(sel, text string) {
		text = strings.TrimSpace(text)
		if text == "" || seen[text] {
			return
		}
		if !priceTextPattern.MatchString(text) {
			return
		}
		cents, err := parsePrice(text)
		if err != nil || cents == 0 {
			return
		}
		seen[text] = true
		candidates = append(candidates, Candidate{
			Selector: sel,
			Text:     text,
			Cents:    cents,
		})
	}

	// Strategy 1: elements whose class or id contains a price keyword
	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		id, _ := s.Attr("id")
		combined := strings.ToLower(class + " " + id)

		for _, kw := range priceKeywords {
			if strings.Contains(combined, kw) {
				text := strings.TrimSpace(s.Text())
				if len(text) > 30 {
					continue
				}
				add(buildSelector(s), text)
				break
			}
		}
	})

	// Strategy 2: any element whose direct text looks like a price
	doc.Find("span, div, p, strong, b, td, li").Each(func(_ int, s *goquery.Selection) {
		text := ""
		s.Contents().Each(func(_ int, c *goquery.Selection) {
			if goquery.NodeName(c) == "#text" {
				text += c.Text()
			}
		})
		add(buildSelector(s), text)
	})

	return candidates, nil
}

// buildSelector constructs a best-effort CSS selector for the given element.
// Priority: id > tag.firstClass > tag
func buildSelector(s *goquery.Selection) string {
	if id, exists := s.Attr("id"); exists && id != "" {
		return "#" + cssEscape(id)
	}
	tag := goquery.NodeName(s)
	if class, exists := s.Attr("class"); exists && class != "" {
		first := strings.Fields(class)[0]
		return tag + "." + cssEscape(first)
	}
	return tag
}

// cssEscape does a minimal escape of characters that are special in CSS selectors.
func cssEscape(s string) string {
	replacer := strings.NewReplacer(
		":", "\\:",
		".", "\\.",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
	)
	return replacer.Replace(s)
}
