package pricectl

import "time"

// ProductOutput is the canonical JSON representation of a product with its
// latest stored price. Used by both the CLI list command and the API.
type ProductOutput struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	PriceCents *int64 `json:"price_cents"`
}

// CheckOutput is the canonical JSON representation of a single check result.
// Used by both the CLI check command and the API.
type CheckOutput struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	PriceCents       int64  `json:"price_cents"`
	OldPriceCents    *int64 `json:"old_price_cents"`
	Changed          bool   `json:"changed"`
	IsNew            bool   `json:"is_new"`
	StructureChanged bool   `json:"structure_changed"`
	Error            string `json:"error,omitempty"`
}

// HistoryEntryOutput is the canonical JSON representation of a single price
// entry, omitting fields that are irrelevant for pipeline consumption.
type HistoryEntryOutput struct {
	PriceCents int64     `json:"price_cents"`
	Timestamp  time.Time `json:"timestamp"`
}

// HistoryOutput is the canonical JSON representation of a product's price history.
type HistoryOutput struct {
	Name    string               `json:"name"`
	Entries []HistoryEntryOutput `json:"entries"`
}
