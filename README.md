# pricewatcher

A primarily command-line tool for monitoring prices on websites. It fetches
configured product pages, extracts prices using CSS selectors and optional
regular expressions, and reports when a price has changed. It also ships a
lightweight web UI that can be launched on demand and viewed in any browser.

## About this project

This tool was written entirely with the assistance of an AI coding agent,
guided throughout by an experienced developer. The process was deliberate and
hands-on: every decision was reviewed, the architecture was shaped with care,
and code quality was treated as a first-class concern throughout — clean
abstractions, a proper storage interface, dependency injection for
testability, and a meaningful test suite.

This is not vibe coding. No code was accepted blindly. The AI was used as a
productive tool, not as a substitute for engineering judgement.

## Installation

Build and install the binary using `make`:

    make build    # builds ./pricewatcher
    make test     # runs the test suite
    make install  # builds and moves the binary to /usr/local/bin

## Configuration

pricewatcher stores its data in `~/.pricewatcher/`:

- `config.json` — the list of products to watch
- `prices.json` — the recorded price history

Both files are created automatically when you add your first product.

A product entry in `config.json` looks like this:

    {
      "name": "Filter Table VST",
      "url": "https://kilohearts.com/products/filter_table",
      "selector": "span.price",
      "regex": "([\d.,]+)"
    }

The `selector` field is a CSS selector that identifies the element containing
the price on the page. The `regex` field is optional — use it when the element
contains additional text around the price that needs to be stripped out.

## Commands

### check

Fetches all configured products, compares the current price with the last
recorded price, and prints a summary. Products whose price has changed are
highlighted. This is the main command you will run periodically.

    pricewatcher check

Pass `--json` to get machine-readable output suitable for use in scripts and
pipelines. Each product is represented as an object in a JSON array:

    pricewatcher check --json

```json
[
  {
    "name": "Filter Table VST",
    "url": "https://kilohearts.com/products/filter_table",
    "price_cents": 4900,
    "old_price_cents": 4900,
    "changed": false,
    "structure_changed": false,
    "is_new": false
  }
]
```

`structure_changed` is set to `true` when the price value is unchanged but the
HTML structure of the price element has changed — this can indicate a sale
overlay or a page redesign and warrants manual verification. `error` is
included as a string field when fetching or parsing a product failed.

### list

Lists all configured products with their most recently recorded price and URL.

    pricewatcher list

Pass `--json` for machine-readable output:

    pricewatcher list --json

```json
[
  {
    "name": "Filter Table VST",
    "url": "https://kilohearts.com/products/filter_table",
    "price_cents": 4900
  }
]
```

`price_cents` is `null` for products that have not been checked yet.

### history

Shows the full price history for a single product, including the direction and
amount of each change. If no name is given, history for all products is shown.

    pricewatcher history "Filter Table VST"

Pass `--json` for machine-readable output:

    pricewatcher history --json
    pricewatcher history --json "Filter Table VST"

```json
[
  {
    "name": "Filter Table VST",
    "entries": [
      { "price_cents": 4900, "timestamp": "2025-04-01T09:00:00Z" },
      { "price_cents": 3900, "timestamp": "2025-04-10T09:00:00Z" }
    ]
  }
]
```

### add

Interactively adds a new product. The tool fetches the page, analyses the HTML
to find elements that look like prices, and presents you with a numbered list
of candidates. You pick one, give the product a name, and pricewatcher writes
the entry to your config file.

    pricewatcher add https://kilohearts.com/products/filter_table

### serve

Starts a local web server on `http://127.0.0.1:8080` and opens a browser-based
UI. The UI shows all configured products with their latest price, lets you
trigger a price check with a button, and displays the full price history for
each product.

    pricewatcher serve

## Running periodically

To check prices automatically, add a cron job. For example, to run every day
at 9:00:

    crontab -e

    0 9 * * * /usr/local/bin/pricewatcher check >> ~/.pricewatcher/check.log 2>&1

For scripted use, `--json` makes it straightforward to pipe the output into
other tools. For example, to list only products whose price dropped:

    pricewatcher check --json | jq '[.[] | select(.changed and .price_cents < .old_price_cents)]'

## Limitations

pricewatcher only works with pages that include the price in the static HTML
response. Pages that load prices dynamically via JavaScript after the initial
page load are not supported.

## Architecture

See [CONTEXT.md](CONTEXT.md) for a detailed record of architectural decisions,
design rationale, and known limitations. It is intended as a reference for
picking up development without losing context.
