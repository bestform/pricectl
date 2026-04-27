package pricectl

import (
	"log"
	"time"
)

// checkResult holds the outcome of checking a single product's price.
type checkResult struct {
	product        Product
	newPrice       int64
	oldPrice       *int64 // nil if no previous price exists
	changed        bool
	rawTextChanged bool // true if price is unchanged but raw extracted text differs
	err            error
}

// fetchFn is the signature for a function that retrieves the current price
// of a product in cents along with the raw extracted text.
// Defined as a type so tests can inject a stub.
type fetchFn func(Product) (int64, string, error)

// isNew reports whether this result represents the first-ever price check for
// the product — no previous entry existed and no fetch error occurred.
func (r checkResult) isNew() bool {
	return r.oldPrice == nil && r.err == nil
}

// checkProduct fetches the current price for a product using the provided
// fetch function, compares it with the stored latest price, persists a new
// entry if the price changed, and returns a checkResult.
//
// ElementHTML change detection: if the price is unchanged but the outerHTML of
// the matched element differs from the stored value, rawTextChanged is set to
// true in the result so callers can warn the user.
func checkProduct(p Product, store Store, fetch fetchFn) checkResult {
	cents, elementHTML, err := fetch(p)
	if err != nil {
		return checkResult{product: p, err: err}
	}

	latest, err := store.LatestPrice(p.Name)
	if err != nil {
		return checkResult{product: p, err: err}
	}

	r := checkResult{product: p, newPrice: cents}

	if latest != nil {
		r.oldPrice = &latest.PriceCents
		r.changed = cents != latest.PriceCents

		if !r.changed && latest.ElementHTML != elementHTML {
			// Same price, but element HTML changed — possible page restructure.
			r.rawTextChanged = true
		}
	}

	// Only store a new entry if the price changed or this is the first check.
	if latest == nil || cents != latest.PriceCents {
		if err := store.AddEntry(p.Name, PriceEntry{
			PriceCents:  cents,
			ElementHTML: elementHTML,
			Timestamp:   time.Now().UTC(),
		}); err != nil {
			log.Printf("warning: failed to store price entry for %s: %v", p.Name, err)
		}
	}

	return r
}
