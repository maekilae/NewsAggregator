package config

import (
	"encoding/json"
	"log/slog"
	"os"
)

type Config struct {
	Providers []ProviderConfig
}

type ProviderConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	return config, nil
}

func (c *Config) LoadProviderConfigs() error {
	entries, err := os.ReadDir("./config/providers/")
	if err != nil {
		slog.Error("Failed to read provider directory", err)
		return err
	}
	for _, entry := range entries {
		value, err := os.ReadFile(entry.Name())
		if err != nil {
			slog.Error("Failed to read provider file", err)
			continue
		}
		var provider ProviderConfig
		err = json.Unmarshal(value, &provider)
		if err != nil {
			slog.Error("Failed to unmarshal provider JSON", err)
			continue
		}
		c.Providers = append(c.Providers, provider)
	}
	return nil
}
