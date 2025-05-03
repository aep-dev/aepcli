package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	APIs map[string]API
}

type API struct {
	Name        string
	OpenAPIPath string
	ServerURL   string
	Headers     []string
	PathPrefix  string
}

func ReadConfigFromFile(file string) (*Config, error) {
	// Check if file exists first
	if _, err := os.Stat(file); os.IsNotExist(err) {
		slog.Debug("Config file does not exist, using default configuration", "file", file)
		return &Config{APIs: make(map[string]API)}, nil
	}

	var c Config
	if _, err := toml.DecodeFile(file, &c); err != nil {
		return nil, fmt.Errorf("unable to decode config file at %v: %v", file, err)
	}
	return &c, nil
}

func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "aepcli"), nil
}

// WriteAPIWithName writes a new API configuration to the specified config file.
func WriteAPIWithName(file string, api API, overwrite bool) error {
	if api.Name == "" {
		return errors.New("api name cannot be empty")
	}

	// Read existing config
	cfg, err := ReadConfigFromFile(file)
	if err != nil {
		// If file doesn't exist yet, initialize new config
		if errors.Is(err, os.ErrNotExist) {
			cfg = &Config{
				APIs: make(map[string]API),
			}
		} else {
			return fmt.Errorf("failed to read existing config: %w", err)
		}
	}

	// Check if API already exists
	if _, exists := cfg.APIs[api.Name]; exists && !overwrite {
		return fmt.Errorf("API with name '%s' already exists. Set --overwrite to true to overwrite", api.Name)
	}

	// Add/update API in config
	cfg.APIs[api.Name] = api

	// Ensure parent directory exists
	parentDir := filepath.Dir(file)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for config file: %w", err)
	}

	// Open file for writing
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	// Encode and write config
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func DefaultConfigFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// ListAPIs returns a slice of all API configurations in the specified config file.
func ListAPIs(file string) ([]API, error) {
	cfg, err := ReadConfigFromFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	apis := make([]API, 0, len(cfg.APIs))
	for _, api := range cfg.APIs {
		apis = append(apis, api)
	}
	return apis, nil
}
