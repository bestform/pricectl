package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		cmdCheck()
	case "list":
		cmdList()
	case "history":
		name := ""
		if len(os.Args) >= 3 {
			name = os.Args[2]
		}
		cmdHistory(name)
	case "serve":
		cmdServe()
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: pricewatcher add <url>")
			os.Exit(1)
		}
		cmdAdd(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("usage: pricewatcher <command>")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  check                                      fetch all products and report price changes")
	fmt.Println("  list                                       list all products with their latest price")
	fmt.Println("  history [name]                             show price history (all products or one by name)")
	fmt.Println("  serve                                      start the web UI on http://127.0.0.1:8080")
	fmt.Println("  add <url>                                  interactively add a new product")
}
