package main

import "time"

// checkResult holds the outcome of checking a single product's price.
type checkResult struct {
	product  Product
	newPrice int64
	oldPrice *int64 // nil if no previous price exists
	changed  bool
	err      error
}

// fetchFn is the signature for a function that retrieves the current price
// of a product in cents. Defined as a type so tests can inject a stub.
type fetchFn func(Product) (int64, error)

// checkProduct fetches the current price for a product using the provided
// fetch function, compares it with the stored latest price, persists a new
// entry if the price changed, and returns a checkResult.
func checkProduct(p Product, store Store, fetch fetchFn) checkResult {
	cents, err := fetch(p)
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
	}

	// Only store a new entry if the price changed or this is the first check
	if latest == nil || cents != latest.PriceCents {
		_ = store.AddEntry(p.Name, PriceEntry{
			PriceCents: cents,
			Timestamp:  time.Now().UTC(),
		})
	}

	return r
}
