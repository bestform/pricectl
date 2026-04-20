package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// jsonStore implements Store using a single JSON file.
type jsonStore struct {
	path string
}

// priceData is the on-disk structure of the prices file.
type priceData struct {
	Prices map[string][]PriceEntry `json:"prices"`
}

// newJSONStore creates a jsonStore backed by ~/.pricectl/prices.json.
func newJSONStore() (*jsonStore, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	return &jsonStore{path: filepath.Join(dir, "prices.json")}, nil
}

func (s *jsonStore) load() (*priceData, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return &priceData{Prices: make(map[string][]PriceEntry)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read prices file: %w", err)
	}
	var pd priceData
	if err := json.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("could not parse prices file: %w", err)
	}
	if pd.Prices == nil {
		pd.Prices = make(map[string][]PriceEntry)
	}
	return &pd, nil
}

func (s *jsonStore) save(pd *priceData) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal prices: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *jsonStore) GetHistory(productName string) ([]PriceEntry, error) {
	pd, err := s.load()
	if err != nil {
		return nil, err
	}
	return pd.Prices[productName], nil
}

func (s *jsonStore) AddEntry(productName string, entry PriceEntry) error {
	pd, err := s.load()
	if err != nil {
		return err
	}
	pd.Prices[productName] = append(pd.Prices[productName], entry)
	return s.save(pd)
}

func (s *jsonStore) LatestPrice(productName string) (*PriceEntry, error) {
	entries, err := s.GetHistory(productName)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	e := entries[len(entries)-1]
	return &e, nil
}

func (s *jsonStore) UpdateLatestElementHTML(productName string, elementHTML string) error {
	pd, err := s.load()
	if err != nil {
		return err
	}
	entries := pd.Prices[productName]
	if len(entries) == 0 {
		return nil
	}
	entries[len(entries)-1].ElementHTML = elementHTML
	pd.Prices[productName] = entries
	return s.save(pd)
}
