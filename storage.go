package main

import "time"

// PriceEntry is a single recorded price at a point in time.
// Price is stored in cents (integer) to avoid floating-point issues.
// ElementHTML is the outerHTML of the matched price element.
// It is used to detect page structure changes even when the parsed price is unchanged.
type PriceEntry struct {
	PriceCents  int64     `json:"price_cents"`
	ElementHTML string    `json:"element_html,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// Store is the interface for reading and writing price history.
// This abstraction allows swapping the backend later (e.g. SQLite).
type Store interface {
	// GetHistory returns all recorded price entries for a product.
	GetHistory(productName string) ([]PriceEntry, error)

	// AddEntry appends a new price entry for a product.
	AddEntry(productName string, entry PriceEntry) error

	// LatestPrice returns the most recent price entry for a product.
	// Returns nil, nil if no entries exist yet.
	LatestPrice(productName string) (*PriceEntry, error)

	// UpdateLatestElementHTML sets the ElementHTML field on the most recent entry
	// for a product. Used to backfill ElementHTML on existing entries that
	// predate the field's introduction.
	UpdateLatestElementHTML(productName string, elementHTML string) error
}
