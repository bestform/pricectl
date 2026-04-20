package main

import (
	"fmt"
	"os"

	"pricectl/internal"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		pricectl.CmdCheck(hasFlag(os.Args[2:], "--json"))
	case "list":
		pricectl.CmdList(hasFlag(os.Args[2:], "--json"))
	case "history":
		args := os.Args[2:]
		jsonOutput := hasFlag(args, "--json")
		name := ""
		for _, a := range args {
			if a != "--json" {
				name = a
				break
			}
		}
		pricectl.CmdHistory(name, jsonOutput)
	case "serve":
		pricectl.CmdServe()
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: pricectl add <url>")
			os.Exit(1)
		}
		pricectl.CmdAdd(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// hasFlag reports whether args contains the given flag string.
func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Println("usage: pricectl <command>")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  check [--json]                             fetch all products and report price changes")
	fmt.Println("  list [--json]                              list all products with their latest price")
	fmt.Println("  history [--json] [name]                    show price history (all products or one by name)")
	fmt.Println("  serve                                      start the web UI on http://127.0.0.1:8080")
	fmt.Println("  add <url>                                  interactively add a new product")
}
