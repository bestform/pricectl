package pricectl

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// historyEntryJSONOutput is the JSON representation of a single price entry,
// omitting fields that are irrelevant for pipeline consumption.
type historyEntryJSONOutput struct {
	PriceCents int64     `json:"price_cents"`
	Timestamp  time.Time `json:"timestamp"`
}

// historyJSONOutput is the JSON representation of a product's price history.
type historyJSONOutput struct {
	Name    string                   `json:"name"`
	Entries []historyEntryJSONOutput `json:"entries"`
}

func CmdHistory(name string, jsonOutput bool) {
	store, err := newJSONStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	var names []string
	if name != "" {
		names = []string{name}
	} else {
		cfg, err := loadConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		for _, p := range cfg.Products {
			names = append(names, p.Name)
		}
	}

	if jsonOutput {
		out := make([]historyJSONOutput, 0, len(names))
		for _, n := range names {
			entries, err := store.GetHistory(n)
			if err != nil {
				continue
			}
			jsonEntries := make([]historyEntryJSONOutput, len(entries))
			for j, e := range entries {
				jsonEntries[j] = historyEntryJSONOutput{PriceCents: e.PriceCents, Timestamp: e.Timestamp}
			}
			out = append(out, historyJSONOutput{Name: n, Entries: jsonEntries})
		}
		writeHistoryJSON(os.Stdout, out)
		return
	}

	for _, n := range names {
		printHistory(store, n)
	}
}

// writeHistoryJSON encodes history for a list of products as a JSON array to w.
func writeHistoryJSON(w io.Writer, items []historyJSONOutput) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(items)
}

func printHistory(store Store, name string) {
	entries, err := store.GetHistory(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return
	}
	if len(entries) == 0 {
		fmt.Printf(bold("── History: %s ")+"\n", name)
		fmt.Println("  no history recorded")
		fmt.Println()
		return
	}

	fmt.Printf(bold("── History: %s ")+"\n", name)
	for i, e := range entries {
		ts := e.Timestamp.Format("2006-01-02 15:04")
		if i == 0 {
			fmt.Printf("  %s  %s\n", ts, formatCents(e.PriceCents))
		} else {
			prev := entries[i-1].PriceCents
			fmt.Printf("  %s  %s  %s %s\n", ts, formatCents(e.PriceCents), priceArrow(prev, e.PriceCents), formatDiff(prev, e.PriceCents))
		}
	}
	fmt.Println()
}
