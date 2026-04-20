package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// --- writeCheckJSON tests ---

func TestWriteCheckJSON_Basic(t *testing.T) {
	oldPrice := int64(3999)
	results := []checkResult{
		{
			product:  Product{Name: "Widget", URL: "https://example.com/widget"},
			newPrice: 2999,
			oldPrice: &oldPrice,
			changed:  true,
		},
	}

	var buf bytes.Buffer
	writeCheckJSON(&buf, results)

	var out []checkJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Name != "Widget" {
		t.Errorf("name = %q, want %q", out[0].Name, "Widget")
	}
	if out[0].PriceCents != 2999 {
		t.Errorf("price_cents = %d, want 2999", out[0].PriceCents)
	}
	if out[0].OldPriceCents == nil || *out[0].OldPriceCents != 3999 {
		t.Errorf("old_price_cents = %v, want 3999", out[0].OldPriceCents)
	}
	if !out[0].Changed {
		t.Errorf("changed should be true")
	}
	if out[0].Error != "" {
		t.Errorf("error should be empty, got %q", out[0].Error)
	}
}

func TestWriteCheckJSON_WithError(t *testing.T) {
	results := []checkResult{
		{
			product: Product{Name: "Broken", URL: "https://example.com/broken"},
			err:     fmt.Errorf("connection refused"),
		},
	}

	var buf bytes.Buffer
	writeCheckJSON(&buf, results)

	var out []checkJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out[0].Error != "connection refused" {
		t.Errorf("error = %q, want %q", out[0].Error, "connection refused")
	}
	if out[0].IsNew {
		t.Errorf("is_new should be false when there is an error")
	}
}

func TestWriteCheckJSON_StructureChanged(t *testing.T) {
	oldPrice := int64(2999)
	results := []checkResult{
		{
			product:        Product{Name: "Widget", URL: "https://example.com"},
			newPrice:       2999,
			oldPrice:       &oldPrice,
			rawTextChanged: true,
		},
	}

	var buf bytes.Buffer
	writeCheckJSON(&buf, results)

	var out []checkJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !out[0].StructureChanged {
		t.Errorf("structure_changed should be true")
	}
	if out[0].Changed {
		t.Errorf("changed should be false when only structure changed")
	}
}

func TestWriteCheckJSON_IsNew(t *testing.T) {
	results := []checkResult{
		{
			product:  Product{Name: "Widget", URL: "https://example.com"},
			newPrice: 2999,
			oldPrice: nil,
		},
	}

	var buf bytes.Buffer
	writeCheckJSON(&buf, results)

	var out []checkJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !out[0].IsNew {
		t.Errorf("is_new should be true when oldPrice is nil and no error")
	}
}

func TestWriteCheckJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	writeCheckJSON(&buf, []checkResult{})

	var out []checkJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty array, got %d elements", len(out))
	}
}

// --- writeListJSON tests ---

func TestWriteListJSON_Basic(t *testing.T) {
	price := int64(1999)
	items := []listJSONOutput{
		{Name: "Widget", URL: "https://example.com/widget", PriceCents: &price},
		{Name: "Gadget", URL: "https://example.com/gadget", PriceCents: nil},
	}

	var buf bytes.Buffer
	writeListJSON(&buf, items)

	var out []listJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	if out[0].Name != "Widget" {
		t.Errorf("name = %q, want %q", out[0].Name, "Widget")
	}
	if out[0].PriceCents == nil || *out[0].PriceCents != 1999 {
		t.Errorf("price_cents = %v, want 1999", out[0].PriceCents)
	}
	if out[1].PriceCents != nil {
		t.Errorf("price_cents should be null for product with no price, got %d", *out[1].PriceCents)
	}
}

func TestWriteListJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	writeListJSON(&buf, []listJSONOutput{})

	var out []listJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty array, got %d elements", len(out))
	}
}

// --- writeHistoryJSON tests ---

func TestWriteHistoryJSON_Basic(t *testing.T) {
	ts := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	items := []historyJSONOutput{
		{
			Name: "Widget",
			Entries: []historyEntryJSONOutput{
				{PriceCents: 3999, Timestamp: ts},
				{PriceCents: 2999, Timestamp: ts.Add(24 * time.Hour)},
			},
		},
	}

	var buf bytes.Buffer
	writeHistoryJSON(&buf, items)

	var out []historyJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 product, got %d", len(out))
	}
	if out[0].Name != "Widget" {
		t.Errorf("name = %q, want %q", out[0].Name, "Widget")
	}
	if len(out[0].Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out[0].Entries))
	}
	if out[0].Entries[0].PriceCents != 3999 {
		t.Errorf("first entry price_cents = %d, want 3999", out[0].Entries[0].PriceCents)
	}
	if out[0].Entries[1].PriceCents != 2999 {
		t.Errorf("second entry price_cents = %d, want 2999", out[0].Entries[1].PriceCents)
	}
}

func TestWriteHistoryJSON_EmptyEntries(t *testing.T) {
	items := []historyJSONOutput{
		{Name: "Widget", Entries: []historyEntryJSONOutput{}},
	}

	var buf bytes.Buffer
	writeHistoryJSON(&buf, items)

	var out []historyJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out[0].Entries) != 0 {
		t.Errorf("expected empty entries array, got %d", len(out[0].Entries))
	}
}

func TestWriteHistoryJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	writeHistoryJSON(&buf, []historyJSONOutput{})

	var out []historyJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty array, got %d elements", len(out))
	}
}

// --- hasFlag tests ---

func TestHasFlag(t *testing.T) {
	if !hasFlag([]string{"--json"}, "--json") {
		t.Error("expected true for ['--json'], '--json'")
	}
	if !hasFlag([]string{"somename", "--json"}, "--json") {
		t.Error("expected true when --json is among other args")
	}
	if hasFlag([]string{"somename"}, "--json") {
		t.Error("expected false when --json is absent")
	}
	if hasFlag([]string{}, "--json") {
		t.Error("expected false for empty args")
	}
}
