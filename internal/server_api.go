package pricectl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// productResponse is the JSON representation of a product with its latest price.
type productResponse struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	PriceCents *int64 `json:"price_cents"`
}

// checkResponse is the JSON representation of a check result.
type checkResponse struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	PriceCents       int64  `json:"price_cents"`
	OldPrice         *int64 `json:"old_price_cents"`
	Changed          bool   `json:"changed"`
	IsNew            bool   `json:"is_new"`
	StructureChanged bool   `json:"structure_changed"`
	Error            string `json:"error,omitempty"`
}

// historyEntryResponse is the JSON representation of a single price entry for the API.
type historyEntryResponse struct {
	PriceCents int64     `json:"price_cents"`
	Timestamp  time.Time `json:"timestamp"`
}

// historyResponse is the JSON representation of history for all products.
type historyResponse struct {
	Name    string                 `json:"name"`
	Entries []historyEntryResponse `json:"entries"`
}

func apiProducts(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadConfig()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	store, err := newJSONStore()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	resp := make([]productResponse, len(cfg.Products))
	for i, p := range cfg.Products {
		pr := productResponse{Name: p.Name, URL: p.URL}
		if latest, err := store.LatestPrice(p.Name); err == nil && latest != nil {
			pr.PriceCents = &latest.PriceCents
		}
		resp[i] = pr
	}
	jsonOK(w, resp)
}

func apiCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	cfg, err := loadConfig()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	store, err := newJSONStore()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	resp := make([]checkResponse, len(cfg.Products))
	for i, p := range cfg.Products {
		result := checkProduct(p, store, fetchPrice)
		cr := checkResponse{
			Name:             p.Name,
			URL:              p.URL,
			PriceCents:       result.newPrice,
			OldPrice:         result.oldPrice,
			Changed:          result.changed,
			IsNew:            result.oldPrice == nil && result.err == nil,
			StructureChanged: result.rawTextChanged,
		}
		if result.err != nil {
			cr.Error = result.err.Error()
		}
		resp[i] = cr
	}
	jsonOK(w, resp)
}

func apiHistory(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadConfig()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	store, err := newJSONStore()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	resp := make([]historyResponse, 0, len(cfg.Products))
	for _, p := range cfg.Products {
		entries, err := store.GetHistory(p.Name)
		if err != nil {
			continue
		}
		respEntries := make([]historyEntryResponse, len(entries))
		for j, e := range entries {
			respEntries[j] = historyEntryResponse{PriceCents: e.PriceCents, Timestamp: e.Timestamp}
		}
		resp = append(resp, historyResponse{Name: p.Name, Entries: respEntries})
	}
	jsonOK(w, resp)
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, err.Error())
}
