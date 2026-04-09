package app

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	CacheEnabled  bool
	PriceCacheTTL time.Duration
	FXCacheTTL    time.Duration
	FinnhubAPIKey string
}

func LoadConfig() (Config, error) {
	// load .env file if present (ignore error if missing)
	_ = godotenv.Load()

	cfg := Config{
		FinnhubAPIKey: os.Getenv("FINNHUB_API_KEY"),
	}

	if cfg.FinnhubAPIKey == "" {
		return Config{}, fmt.Errorf("FINNHUB_API_KEY is required")
	}

	return cfg, nil
}
