package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppConfig struct {
	Host     string `json:"host"`
	Interval int    `json:"interval"`
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "flow")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "config.json"), nil
}

func loadConfig() AppConfig {
	path, err := getConfigPath()
	var cfg AppConfig
	if err == nil {
		b, err := os.ReadFile(path)
		if err == nil {
			json.Unmarshal(b, &cfg)
		}
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 1 // default to 1 second
	}
	return cfg
}

func saveConfig(cfg AppConfig) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	b, _ := json.Marshal(cfg)
	return os.WriteFile(path, b, 0644)
}
