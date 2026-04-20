package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// listJSONOutput is the JSON representation of a single product in list output.
type listJSONOutput struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	PriceCents *int64 `json:"price_cents"`
}

func cmdList(jsonOutput bool) {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(cfg.Products) == 0 {
		if jsonOutput {
			fmt.Fprintln(os.Stdout, "[]")
		} else {
			fmt.Println("no products configured")
			fmt.Printf("add products to %s\n", mustConfigPath())
		}
		return
	}

	store, err := newJSONStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if jsonOutput {
		items := make([]listJSONOutput, len(cfg.Products))
		for i, p := range cfg.Products {
			o := listJSONOutput{Name: p.Name, URL: p.URL}
			if latest, err := store.LatestPrice(p.Name); err == nil && latest != nil {
				o.PriceCents = &latest.PriceCents
			}
			items[i] = o
		}
		writeListJSON(os.Stdout, items)
		return
	}

	fmt.Println(bold("── Products ──────────────────────────────────────"))
	for _, p := range cfg.Products {
		latest, err := store.LatestPrice(p.Name)
		if err != nil || latest == nil {
			fmt.Printf("  %-40s %s  %s\n", p.Name, yellow("no price"), p.URL)
			continue
		}
		fmt.Printf("  %-40s %s  %s\n", p.Name, formatCents(latest.PriceCents), p.URL)
	}
}

// writeListJSON encodes a list of products as a JSON array to w.
func writeListJSON(w io.Writer, items []listJSONOutput) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(items)
}
