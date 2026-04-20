package pricectl

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func CmdAdd(url string) {
	fmt.Printf("Loading %s ...\n", url)
	doc, err := fetchDoc(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading page:", err)
		os.Exit(1)
	}

	fmt.Println("Searching for price candidates...")
	candidates, err := FindPriceCandidates(doc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error analysing page:", err)
		os.Exit(1)
	}

	if len(candidates) == 0 {
		fmt.Println(yellow("No price candidates found automatically."))
		fmt.Println()
		fmt.Println("You can add the product manually by editing ~/.pricectl/config.json")
		fmt.Println("and appending an entry like this:")
		fmt.Println()
		fmt.Println(`  {`)
		fmt.Println(`    "name": "My Product",`)
		fmt.Printf(`    "url": "%s",`+"\n", url)
		fmt.Println(`    "selector": "span.price",`)
		fmt.Println(`    "regex": "([\\d.,]+)"`)
		fmt.Println(`  }`)
		fmt.Println()
		fmt.Println("Use your browser's developer tools to find the right CSS selector.")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(bold("── Price candidates ──────────────────────────────"))
	for i, c := range candidates {
		fmt.Printf("  %s  %-14s  %s\n",
			bold(fmt.Sprintf("[%d]", i+1)),
			formatCents(c.Cents),
			c.Selector,
		)
	}

	fmt.Println()
	reader := bufio.NewReader(os.Stdin)

	var chosen *Candidate
	for {
		fmt.Printf("Which candidate do you want to use? [1-%d, or 'q' to cancel]: ", len(candidates))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "q" || input == "Q" {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(candidates) {
			fmt.Printf("Please enter a number between 1 and %d.\n", len(candidates))
			continue
		}
		c := candidates[n-1]
		chosen = &c
		break
	}

	fmt.Printf("\nChosen: %s → %s\n", chosen.Selector, formatCents(chosen.Cents))

	fmt.Print("Product name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Fprintln(os.Stderr, "name must not be empty")
		os.Exit(1)
	}

	product := Product{
		Name:     name,
		URL:      url,
		Selector: chosen.Selector,
	}

	suggestedRegex := suggestRegex(chosen.Text)
	if suggestedRegex != "" {
		fmt.Printf("Suggested regex (to extract price from %q): %s\n", chosen.Text, suggestedRegex)
		fmt.Print("Use this regex? [Y/n]: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "" || answer == "y" {
			product.Regex = suggestedRegex
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading config:", err)
		os.Exit(1)
	}

	for _, p := range cfg.Products {
		if p.Name == name {
			fmt.Fprintf(os.Stderr, "a product named %q already exists\n", name)
			os.Exit(1)
		}
	}

	cfg.Products = append(cfg.Products, product)

	if err := saveConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error saving config:", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s Product %q added.\n", green("✓"), name)
	fmt.Printf("  Selector: %s\n", product.Selector)
	if product.Regex != "" {
		fmt.Printf("  Regex:    %s\n", product.Regex)
	}
	fmt.Printf("\nRun 'pricectl check' to fetch the current price.\n")
}

// suggestRegex returns a capture-group regex if the text contains extra
// characters around the price number, otherwise returns "".
func suggestRegex(text string) string {
	text = strings.TrimSpace(text)
	barePrice := regexp.MustCompile(`^[€$£¥\s]*[\d.,]+[€$£¥\s]*$`)
	if barePrice.MatchString(text) {
		return ""
	}
	return `([\d.,]+)`
}
