package pricectl

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// apiServer holds shared server-lifetime dependencies for the API handlers.
// Constructing a single instance in CmdServe ensures the Store's mutex
// serialises all concurrent writes across parallel HTTP requests.
type apiServer struct {
	store Store
}

func (s *apiServer) apiProducts(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadConfig()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	resp := make([]ProductOutput, len(cfg.Products))
	for i, p := range cfg.Products {
		o := ProductOutput{Name: p.Name, URL: p.URL}
		if latest, err := s.store.LatestPrice(p.Name); err == nil && latest != nil {
			o.PriceCents = &latest.PriceCents
		}
		resp[i] = o
	}
	jsonOK(w, resp)
}

func (s *apiServer) apiCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	cfg, err := loadConfig()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	// If a name query parameter is given, check only that product.
	if name := r.URL.Query().Get("name"); name != "" {
		var found *Product
		for i := range cfg.Products {
			if cfg.Products[i].Name == name {
				found = &cfg.Products[i]
				break
			}
		}
		if found == nil {
			jsonError(w, fmt.Errorf("product %q not found", name), http.StatusNotFound)
			return
		}
		result := checkProduct(*found, s.store, fetchPrice)
		o := CheckOutput{
			Name:             found.Name,
			URL:              found.URL,
			PriceCents:       result.newPrice,
			OldPriceCents:    result.oldPrice,
			Changed:          result.changed,
			IsNew:            result.isNew(),
			StructureChanged: result.rawTextChanged,
		}
		if result.err != nil {
			o.Error = result.err.Error()
		}
		jsonOK(w, o)
		return
	}

	resp := make([]CheckOutput, len(cfg.Products))
	for i, p := range cfg.Products {
		result := checkProduct(p, s.store, fetchPrice)
		o := CheckOutput{
			Name:             p.Name,
			URL:              p.URL,
			PriceCents:       result.newPrice,
			OldPriceCents:    result.oldPrice,
			Changed:          result.changed,
			IsNew:            result.isNew(),
			StructureChanged: result.rawTextChanged,
		}
		if result.err != nil {
			o.Error = result.err.Error()
		}
		resp[i] = o
	}
	jsonOK(w, resp)
}

func (s *apiServer) apiHistory(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadConfig()
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
		return
	}

	resp := make([]HistoryOutput, 0, len(cfg.Products))
	for _, p := range cfg.Products {
		entries, err := s.store.GetHistory(p.Name)
		if err != nil {
			continue
		}
		respEntries := make([]HistoryEntryOutput, len(entries))
		for j, e := range entries {
			respEntries[j] = HistoryEntryOutput{PriceCents: e.PriceCents, Timestamp: e.Timestamp}
		}
		resp = append(resp, HistoryOutput{Name: p.Name, Entries: respEntries})
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
