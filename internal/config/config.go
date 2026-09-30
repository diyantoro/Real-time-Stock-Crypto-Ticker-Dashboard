package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProviderURL string   `yaml:"provider_url"`
	Symbols     []string `yaml:"symbols"`
}

func Default() Config {
	return Config{
		ProviderURL: "wss://stream.binance.com:9443",
		Symbols:     []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDT"},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return applyEnv(cfg), nil
			}
			return Config{}, fmt.Errorf("baca config %s: %w", path, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse config %s: %w", path, err)
		}
	}
	cfg = applyEnv(cfg)
	if cfg.ProviderURL == "" {
		return Config{}, fmt.Errorf("provider_url kosong")
	}
	if len(cfg.Symbols) == 0 {
		return Config{}, fmt.Errorf("symbols kosong")
	}
	return cfg, nil
}

func applyEnv(cfg Config) Config {
	if v := os.Getenv("PROVIDER_URL"); v != "" {
		cfg.ProviderURL = v
	}
	if v := os.Getenv("SYMBOLS"); v != "" {
		parts := strings.Split(v, ",")
		var out []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, strings.ToUpper(p))
			}
		}
		if len(out) > 0 {
			cfg.Symbols = out
		}
	}
	return cfg
}
