package app

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	PriceProvider string
	FXProvider    string

	FinnhubAPIKey string

	CacheEnabled  bool
	PriceCacheTTL time.Duration
	FXCacheTTL    time.Duration

	LLMEnabled  bool
	LLMProvider string
	LLMBaseURL  string
	LLMModel    string
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		PriceProvider: getenvDefault("PRICE_PROVIDER", "static"),
		FXProvider:    getenvDefault("FX_PROVIDER", "static"),
		FinnhubAPIKey: os.Getenv("FINNHUB_API_KEY"),
		CacheEnabled:  getenvDefault("CACHE_ENABLED", "true") == "true",

		LLMEnabled:  getenvDefault("LLM_ENABLED", "false") == "true",
		LLMProvider: getenvDefault("LLM_PROVIDER", "ollama"),
		LLMBaseURL:  getenvDefault("LLM_BASE_URL", "http://localhost:11434"),
		LLMModel:    getenvDefault("LLM_MODEL", "llama3.1"),
	}

	var err error

	cfg.PriceCacheTTL, err = durationEnv("CACHE_PRICE_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("CACHE_PRICE_TTL: %w", err)
	}

	cfg.FXCacheTTL, err = durationEnv("CACHE_FX_TTL", 12*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("CACHE_FX_TTL: %w", err)
	}

	if cfg.PriceProvider == "finnhub" || cfg.FXProvider == "finnhub" {
		if cfg.FinnhubAPIKey == "" {
			return Config{}, fmt.Errorf("FINNHUB_API_KEY is required when using finnhub providers")
		}
	}

	if cfg.LLMEnabled {
		switch cfg.LLMProvider {
		case "ollama":
			if cfg.LLMBaseURL == "" {
				return Config{}, fmt.Errorf("LLM_BASE_URL is required when LLM is enabled")
			}
			if cfg.LLMModel == "" {
				return Config{}, fmt.Errorf("LLM_MODEL is required when LLM is enabled")
			}
		default:
			return Config{}, fmt.Errorf("unsupported LLM_PROVIDER %q", cfg.LLMProvider)
		}
	}

	return cfg, nil
}

func getenvDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return time.ParseDuration(v)
}
