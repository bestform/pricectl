package pricectl

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// fetchPrice fetches the URL of a product, applies the CSS selector and
// optional regex, and returns the price in cents and the outerHTML of the
// matched element.
func fetchPrice(p Product) (int64, string, error) {
	doc, err := fetchDoc(p.URL)
	if err != nil {
		return 0, "", err
	}
	return extractPrice(doc, p.Selector, p.Regex)
}

// fetchDoc performs an HTTP GET and returns a parsed goquery document.
// A browser-like User-Agent is set to avoid simple bot-detection blocks.
func fetchDoc(url string) (*goquery.Document, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse HTML: %w", err)
	}
	return doc, nil
}

// extractPrice finds elements matching selector in the document and extracts
// a price in cents from their text content.
// If regex is non-empty, it is applied to the element text to isolate the number.
// If the first match is not parseable as a price, subsequent matches are tried.
// Returns the price in cents and the outerHTML of the matched element.
func extractPrice(doc *goquery.Document, selector, pattern string) (int64, string, error) {
	matches := doc.Find(selector)
	if matches.Length() == 0 {
		return 0, "", fmt.Errorf("selector %q matched no elements", selector)
	}

	var lastErr error
	var result int64
	var elementHTML string
	found := false

	matches.EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		text := strings.TrimSpace(sel.Text())
		if text == "" {
			lastErr = fmt.Errorf("selector %q matched an empty element", selector)
			return true // try next
		}

		candidate := text
		if pattern != "" {
			re, err := regexp.Compile(pattern)
			if err != nil {
				lastErr = fmt.Errorf("invalid regex %q: %w", pattern, err)
				return false // regex error is fatal
			}
			m := re.FindStringSubmatch(text)
			if len(m) < 2 {
				lastErr = fmt.Errorf("regex %q found no match in %q", pattern, text)
				return true // try next element
			}
			candidate = m[1]
		}

		cents, err := parsePrice(candidate)
		if err != nil || cents == 0 {
			lastErr = fmt.Errorf("could not parse %q as price", candidate)
			return true // try next element
		}

		html, err := goquery.OuterHtml(sel)
		if err != nil {
			lastErr = fmt.Errorf("could not get outerHTML: %w", err)
			return true // try next element
		}

		result = cents
		elementHTML = html
		found = true
		return false // stop iterating
	})

	if found {
		return result, elementHTML, nil
	}
	if lastErr != nil {
		return 0, "", fmt.Errorf("selector %q matched %d element(s) but none contained a valid price (last error: %w)", selector, matches.Length(), lastErr)
	}
	return 0, "", fmt.Errorf("selector %q matched no parseable price", selector)
}

// parsePrice converts a price string to cents.
// Handles formats like "499", "29,99", "29.99", "1.299,00"
func parsePrice(s string) (int64, error) {
	s = strings.TrimSpace(s)

	// Remove currency symbols and spaces
	s = strings.ReplaceAll(s, "€", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)

	hasDot := strings.Contains(s, ".")
	hasComma := strings.Contains(s, ",")

	var intPart, fracPart string

	switch {
	case hasDot && hasComma:
		// Determine which is decimal by position of last separator
		lastDot := strings.LastIndex(s, ".")
		lastComma := strings.LastIndex(s, ",")
		if lastComma > lastDot {
			// e.g. "1.299,00" — dot is thousands, comma is decimal
			s = strings.ReplaceAll(s, ".", "")
			parts := strings.SplitN(s, ",", 2)
			intPart = parts[0]
			fracPart = parts[1]
		} else {
			// e.g. "1,299.00" — comma is thousands, dot is decimal
			s = strings.ReplaceAll(s, ",", "")
			parts := strings.SplitN(s, ".", 2)
			intPart = parts[0]
			fracPart = parts[1]
		}
	case hasComma:
		// e.g. "29,99" — comma is decimal separator
		parts := strings.SplitN(s, ",", 2)
		intPart = parts[0]
		fracPart = parts[1]
	case hasDot:
		// e.g. "29.99" — dot is decimal separator
		parts := strings.SplitN(s, ".", 2)
		intPart = parts[0]
		fracPart = parts[1]
	default:
		// e.g. "499" — whole number, no cents
		intPart = s
		fracPart = "00"
	}

	// Normalize fracPart to exactly 2 digits
	switch len(fracPart) {
	case 0:
		fracPart = "00"
	case 1:
		fracPart = fracPart + "0"
	default:
		fracPart = fracPart[:2]
	}

	i, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse integer part %q: %w", intPart, err)
	}
	f, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse fractional part %q: %w", fracPart, err)
	}

	return i*100 + f, nil
}
