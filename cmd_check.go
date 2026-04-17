package main

import (
	"fmt"
	"os"
)

func cmdCheck() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(cfg.Products) == 0 {
		fmt.Println("no products configured")
		fmt.Printf("add products to %s\n", mustConfigPath())
		return
	}

	store, err := newJSONStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	results := make([]checkResult, len(cfg.Products))
	for i, p := range cfg.Products {
		fmt.Printf("fetching %-40s ", p.Name+"...")
		r := checkProduct(p, store, fetchPrice)
		if r.err != nil {
			fmt.Println(yellow("ERROR"))
		} else {
			fmt.Println("OK")
		}
		results[i] = r
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
		} else {
			fmt.Printf("  %-40s %s  %s\n", r.product.Name, price, priceArrow(*r.oldPrice, r.newPrice))
		}
	}
}
