package config

import (
	"fmt"
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

// ReadConfig reads the configuration from the default configuration file.
func ReadConfig() (*Config, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	return ReadConfigFromFile(filepath.Join(configDir, "config.toml"))
}

func ReadConfigFromFile(filename string) (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(filename, &c); err != nil {
		return nil, fmt.Errorf("unable to decode config file at %v: %v", filename, err)
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
