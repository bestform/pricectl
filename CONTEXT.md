# pricewatcher — Project Context

This document captures the product decisions, architectural choices, and
reasoning behind them. It is intended as a reference for picking up development
in the future without losing context.

---

## What the product is

pricewatcher is a personal tool for tracking prices on websites. It runs
locally on the developer's own machine. The user configures a list of products
— each with a URL, a CSS selector, and an optional regular expression — and the
tool periodically fetches those pages, extracts the price, and reports changes.

The tool is primarily a CLI, designed to be invoked from a terminal or via a
cron job. All CLI commands that produce output support a `--json` flag for
machine-readable output suitable for scripts and pipelines. It also ships an
optional web UI for a more visual overview.

---

## Core decisions and their rationale

### Language: Go

Go was chosen for its simplicity, fast compile times, and the fact that it
produces a single self-contained binary with no runtime dependencies. This
makes installation and distribution trivial for a personal tool.

### Data storage: JSON files in `~/.pricewatcher/`

No database. Two JSON files are sufficient for the scale of this tool
(`config.json` for product configuration, `prices.json` for price history).
The storage layer is hidden behind a `Store` interface, so switching to SQLite
or another backend in the future requires no changes to the rest of the code.

Prices are stored as integers in cents to avoid floating-point precision
issues entirely.

### Price extraction: CSS selector + optional regex

Each product is configured with a CSS selector that identifies the element
containing the price. An optional regular expression can be used to strip
surrounding text if necessary. Only static HTML is parsed — no headless browser
or JavaScript execution. This is a deliberate limitation: supporting dynamic
pages would add significant complexity and dependencies for what is, in
practice, rarely needed (most shops include the price in the initial HTML).

### Flat file structure, no subpackages

All Go files live in a single `main` package in the root directory, grouped by
a naming convention (`cmd_*.go` for commands, `store*.go` for storage, etc.).
Splitting a tool of this size into subpackages would be over-engineering and
would add friction without meaningful benefit.

### Dependency injection for testability

The `checkProduct` function accepts a `fetchFn` parameter instead of calling
the HTTP layer directly. This allows tests to inject a stub and run without
making real network requests. The same pattern applies to the `Store` interface.

### The `inspect` command was removed

An `inspect` command was built early on to test a CSS selector against a URL
without saving anything. It was removed because the `add` command's interactive
heuristic covers the common case well enough, and the extra command added
surface area without enough ongoing value. If manual inspection becomes
necessary in the future, a more purpose-built approach can be added then.

### Web UI: embedded HTML, no framework

The web UI is a single HTML file embedded directly into the binary via Go's
`embed` package. It uses plain HTML, CSS, and vanilla JavaScript — no build
step, no npm, no external framework. The server exposes three JSON endpoints
(`/api/products`, `/api/check`, `/api/history`) that the frontend consumes.
`/api/check` is restricted to POST requests to prevent accidental triggering
(e.g. by browser prefetch).

The UI was added as a convenience layer on top of the CLI. The CLI remains the
primary interface and is the one recommended for cron-based automation.

When check results are displayed, the product list is hidden because the check
results already contain the same information — showing both would be redundant.

### Structure change detection

When `checkProduct` finds that the price is unchanged but the outerHTML of the
matched element differs from the stored value, it flags this as a structure
change (`structure_changed` in JSON output). This can indicate a sale overlay,
a redesigned page, or a selector that now matches a different element — all
cases where the user should verify the price manually. The raw element HTML is
stored in `PriceEntry.ElementHTML` and is deliberately excluded from all JSON
output (both CLI and API) since it is an internal implementation detail.

---

## What exists today

| File | Purpose |
|---|---|
| `main.go` | Command routing only |
| `cmd_check.go` | `check` command |
| `cmd_list.go` | `list` command |
| `cmd_history.go` | `history` command |
| `cmd_add.go` | `add` command with interactive heuristic |
| `cmd_serve.go` | `serve` command, starts HTTP server, embeds UI |
| `server_api.go` | JSON API handlers for the web UI |
| `ui/index.html` | Web UI (embedded into binary) |
| `config.go` | Product config model, load/save/path helpers |
| `storage.go` | `Store` interface and `PriceEntry` type |
| `storage_json.go` | JSON-backed `Store` implementation |
| `checker.go` | `checkProduct` with injected fetch function |
| `fetcher.go` | HTTP fetch and price extraction logic |
| `heuristic.go` | `FindPriceCandidates` for the `add` command |
| `output.go` | Formatting helpers (colors, cents, diffs) |
| `Makefile` | `build`, `test`, `install`, `serve` targets |

Test files exist for: `checker`, `fetcher` (parsePrice, extractPrice),
`heuristic`, `output`, `json_output` (all JSON write functions), and `main`
(hasFlag utility).

---

## Known limitations and future considerations

- **No JavaScript support.** Pages that load prices dynamically after the
  initial HTML response are not supported. Adding headless browser support
  (e.g. via `chromedp`) would be the natural next step if this becomes a
  blocker.

- **No notifications.** The tool reports changes on stdout (CLI) or in the web
  UI, but does not send emails, push notifications, or webhook calls. This
  would be a useful addition for unattended cron runs.

- **Single store file.** All price history is stored in one `prices.json` file.
  For a large number of products with long histories this could become slow to
  read and write. Switching to SQLite via the existing `Store` interface would
  solve this cleanly.

- **No authentication on the web UI.** The server binds to `127.0.0.1` only,
  so it is not reachable from outside the local machine. If the tool were ever
  deployed on a server, authentication would need to be added.
