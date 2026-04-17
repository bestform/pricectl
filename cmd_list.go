package main

import (
	"fmt"
	"os"
)

func cmdList() {
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
