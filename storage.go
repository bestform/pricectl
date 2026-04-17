package main

import "time"

// PriceEntry is a single recorded price at a point in time.
// Price is stored in cents (integer) to avoid floating-point issues.
type PriceEntry struct {
	PriceCents int64     `json:"price_cents"`
	Timestamp  time.Time `json:"timestamp"`
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
}
