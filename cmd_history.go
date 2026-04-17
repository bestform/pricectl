package main

import (
	"fmt"
	"os"
)

func cmdHistory(name string) {
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

	for _, n := range names {
		printHistory(store, n)
	}
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
