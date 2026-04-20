package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Product represents a single product to watch.
type Product struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Selector string `json:"selector"`
	Regex    string `json:"regex,omitempty"`
}

// Config holds all configured products.
type Config struct {
	Products []Product `json:"products"`
}

// configDir returns the path to ~/.pricectl.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".pricectl"), nil
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// loadConfig reads the config file. If the file does not exist, an empty
// Config is returned without error.
func loadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}
	return &cfg, nil
}

// ensureConfigDir creates ~/.pricectl if it does not exist.
func ensureConfigDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// saveConfig writes the config to ~/.pricectl/config.json.
func saveConfig(cfg *Config) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
