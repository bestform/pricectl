package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// checkJSONOutput is the JSON representation of a single check result.
type checkJSONOutput struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	PriceCents       int64  `json:"price_cents"`
	OldPriceCents    *int64 `json:"old_price_cents"`
	Changed          bool   `json:"changed"`
	StructureChanged bool   `json:"structure_changed"`
	IsNew            bool   `json:"is_new"`
	Error            string `json:"error,omitempty"`
}

func cmdCheck(jsonOutput bool) {
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

	results := make([]checkResult, len(cfg.Products))
	for i, p := range cfg.Products {
		if !jsonOutput {
			fmt.Printf("fetching %-40s ", p.Name+"...")
		}
		r := checkProduct(p, store, fetchPrice)
		if !jsonOutput {
			if r.err != nil {
				fmt.Println(yellow("ERROR"))
			} else {
				fmt.Println("OK")
			}
		}
		results[i] = r
	}

	if jsonOutput {
		writeCheckJSON(os.Stdout, results)
		return
	}

	fmt.Println()
	fmt.Println(bold("── Results ───────────────────────────────────────"))
	for _, r := range results {
		if r.err != nil {
			fmt.Printf("  %-40s %s\n", r.product.Name, red("ERROR: "+r.err.Error()))
			continue
		}

		price := formatCents(r.newPrice)

		if r.oldPrice == nil {
			fmt.Printf("  %-40s %s  %s\n", r.product.Name, price, yellow("(new)"))
		} else if r.changed {
			arrow := priceArrow(*r.oldPrice, r.newPrice)
			diff := formatDiff(*r.oldPrice, r.newPrice)
			old := formatCents(*r.oldPrice)
			fmt.Printf("  %-40s %s  %s  (was %s, %s)\n",
				bold(r.product.Name), bold(price), arrow, old, diff)
		} else if r.rawTextChanged {
			fmt.Printf("  %-40s %s  %s  %s\n", r.product.Name, price,
				priceArrow(*r.oldPrice, r.newPrice),
				yellow("⚠ page structure changed — verify price manually"))
		} else {
			fmt.Printf("  %-40s %s  %s\n", r.product.Name, price, priceArrow(*r.oldPrice, r.newPrice))
		}
	}
}

// writeCheckJSON encodes check results as a JSON array to w.
func writeCheckJSON(w io.Writer, results []checkResult) {
	out := make([]checkJSONOutput, len(results))
	for i, r := range results {
		o := checkJSONOutput{
			Name:             r.product.Name,
			URL:              r.product.URL,
			PriceCents:       r.newPrice,
			OldPriceCents:    r.oldPrice,
			Changed:          r.changed,
			StructureChanged: r.rawTextChanged,
			IsNew:            r.oldPrice == nil && r.err == nil,
		}
		if r.err != nil {
			o.Error = r.err.Error()
		}
		out[i] = o
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}
