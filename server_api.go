package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// productResponse is the JSON representation of a product with its latest price.
type productResponse struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	PriceCents *int64 `json:"price_cents"`
}

// checkResponse is the JSON representation of a check result.
type checkResponse struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	PriceCents     int64  `json:"price_cents"`
	OldPrice       *int64 `json:"old_price_cents"`
	Changed        bool   `json:"changed"`
	IsNew          bool   `json:"is_new"`
	RawTextChanged bool   `json:"raw_text_changed"`
	Error          string `json:"error,omitempty"`
}

// historyResponse is the JSON representation of history for all products.
type historyResponse struct {
	Name    string       `json:"name"`
	Entries []PriceEntry `json:"entries"`
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
			Name:           p.Name,
			URL:            p.URL,
			PriceCents:     result.newPrice,
			OldPrice:       result.oldPrice,
			Changed:        result.changed,
			IsNew:          result.oldPrice == nil && result.err == nil,
			RawTextChanged: result.rawTextChanged,
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
		resp = append(resp, historyResponse{Name: p.Name, Entries: entries})
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
