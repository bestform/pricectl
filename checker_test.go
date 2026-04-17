package main

import (
	"fmt"
	"testing"
	"time"
)

// memStore is an in-memory Store implementation used in tests.
type memStore struct {
	entries map[string][]PriceEntry
}

func newMemStore() *memStore {
	return &memStore{entries: make(map[string][]PriceEntry)}
}

func (m *memStore) GetHistory(name string) ([]PriceEntry, error) {
	return m.entries[name], nil
}

func (m *memStore) AddEntry(name string, e PriceEntry) error {
	m.entries[name] = append(m.entries[name], e)
	return nil
}

func (m *memStore) LatestPrice(name string) (*PriceEntry, error) {
	entries := m.entries[name]
	if len(entries) == 0 {
		return nil, nil
	}
	e := entries[len(entries)-1]
	return &e, nil
}

// stubFetch returns a fetchFn that always returns the given cents value.
func stubFetch(cents int64) fetchFn {
	return func(p Product) (int64, error) {
		return cents, nil
	}
}

// failFetch returns a fetchFn that always returns an error.
func failFetch(msg string) fetchFn {
	return func(p Product) (int64, error) {
		return 0, fmt.Errorf("%s", msg)
	}
}

var testProduct = Product{Name: "Test Product", URL: "https://example.com", Selector: ".price"}

func TestCheckProduct_FirstCheck(t *testing.T) {
	store := newMemStore()
	r := checkProduct(testProduct, store, stubFetch(2999))

	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
	if r.newPrice != 2999 {
		t.Errorf("newPrice = %d, want 2999", r.newPrice)
	}
	if r.oldPrice != nil {
		t.Errorf("oldPrice should be nil on first check")
	}
	if r.changed {
		t.Errorf("changed should be false on first check")
	}

	// Entry must have been stored
	entries, _ := store.GetHistory(testProduct.Name)
	if len(entries) != 1 {
		t.Errorf("expected 1 stored entry, got %d", len(entries))
	}
	if entries[0].PriceCents != 2999 {
		t.Errorf("stored price = %d, want 2999", entries[0].PriceCents)
	}
}

func TestCheckProduct_PriceUnchanged(t *testing.T) {
	store := newMemStore()
	_ = store.AddEntry(testProduct.Name, PriceEntry{PriceCents: 2999, Timestamp: time.Now()})

	r := checkProduct(testProduct, store, stubFetch(2999))

	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
	if r.changed {
		t.Errorf("changed should be false when price is the same")
	}
	if *r.oldPrice != 2999 {
		t.Errorf("oldPrice = %d, want 2999", *r.oldPrice)
	}

	// No new entry should have been added
	entries, _ := store.GetHistory(testProduct.Name)
	if len(entries) != 1 {
		t.Errorf("expected 1 stored entry (no new one), got %d", len(entries))
	}
}

func TestCheckProduct_PriceDropped(t *testing.T) {
	store := newMemStore()
	_ = store.AddEntry(testProduct.Name, PriceEntry{PriceCents: 4999, Timestamp: time.Now()})

	r := checkProduct(testProduct, store, stubFetch(2999))

	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
	if !r.changed {
		t.Errorf("changed should be true when price dropped")
	}
	if *r.oldPrice != 4999 {
		t.Errorf("oldPrice = %d, want 4999", *r.oldPrice)
	}
	if r.newPrice != 2999 {
		t.Errorf("newPrice = %d, want 2999", r.newPrice)
	}

	// New entry must have been stored
	entries, _ := store.GetHistory(testProduct.Name)
	if len(entries) != 2 {
		t.Errorf("expected 2 stored entries, got %d", len(entries))
	}
}

func TestCheckProduct_PriceRose(t *testing.T) {
	store := newMemStore()
	_ = store.AddEntry(testProduct.Name, PriceEntry{PriceCents: 2999, Timestamp: time.Now()})

	r := checkProduct(testProduct, store, stubFetch(4999))

	if !r.changed {
		t.Errorf("changed should be true when price rose")
	}
	if *r.oldPrice != 2999 {
		t.Errorf("oldPrice = %d, want 2999", *r.oldPrice)
	}
	if r.newPrice != 4999 {
		t.Errorf("newPrice = %d, want 4999", r.newPrice)
	}
}

func TestCheckProduct_FetchError(t *testing.T) {
	store := newMemStore()
	r := checkProduct(testProduct, store, failFetch("network timeout"))

	if r.err == nil {
		t.Fatal("expected error, got nil")
	}

	// Nothing should have been stored
	entries, _ := store.GetHistory(testProduct.Name)
	if len(entries) != 0 {
		t.Errorf("expected no stored entries on fetch error, got %d", len(entries))
	}
}
