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
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		PriceProvider: getenvDefault("PRICE_PROVIDER", "static"),
		FXProvider:    getenvDefault("FX_PROVIDER", "static"),
		FinnhubAPIKey: os.Getenv("FINNHUB_API_KEY"),
		CacheEnabled:  getenvDefault("CACHE_ENABLED", "true") == "true",
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
