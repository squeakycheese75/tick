package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/squeakycheese75/tick/internal/appdir"
)

type Config struct {
	EquityPriceProvider string
	CryptoPriceProvider string
	FXProvider          string
	NewsProvider        string

	FinnhubAPIKey    string
	NewsAPIOrgAPIKey string

	CacheEnabled  bool
	PriceCacheTTL time.Duration
	FXCacheTTL    time.Duration

	LLMEnabled  bool
	LLMProvider string
	LLMBaseURL  string
	LLMModel    string
}

func LoadConfig() (Config, error) {
	configPath, err := appdir.ConfigPath()
	if err == nil {
		if err := godotenv.Load(configPath); err != nil && !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("load home config: %w", err)
		}
	}

	if err := godotenv.Overload(".env"); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load local .env: %w", err)
	}

	cfg := Config{
		EquityPriceProvider: getenvDefault("EQUITY_PRICE_PROVIDER", "static"),
		CryptoPriceProvider: getenvDefault("CRYPTO_PRICE_PROVIDER", "static"),
		FXProvider:          getenvDefault("FX_PROVIDER", "static"),
		NewsProvider:        getenvDefault("NEWS_PROVIDER", "static"),
		FinnhubAPIKey:       os.Getenv("FINNHUB_API_KEY"),
		NewsAPIOrgAPIKey:    os.Getenv("NEWSAPIORG_API_KEY"),
		CacheEnabled:        getenvDefault("CACHE_ENABLED", "true") == "true",

		LLMEnabled:  getenvDefault("LLM_ENABLED", "false") == "true",
		LLMProvider: getenvDefault("LLM_PROVIDER", "ollama"),
		LLMBaseURL:  getenvDefault("LLM_BASE_URL", "http://localhost:11434"),
		LLMModel:    getenvDefault("LLM_MODEL", "llama3.1"),
	}

	cfg.PriceCacheTTL, err = durationEnv("CACHE_PRICE_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("CACHE_PRICE_TTL: %w", err)
	}

	cfg.FXCacheTTL, err = durationEnv("CACHE_FX_TTL", 12*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("CACHE_FX_TTL: %w", err)
	}

	return cfg, nil
}

func (c Config) Validate() error {
	switch c.EquityPriceProvider {
	case "static", "finnhub":
	default:
		return fmt.Errorf("unsupported EQUITY_PRICE_PROVIDER %q", c.EquityPriceProvider)
	}

	switch c.CryptoPriceProvider {
	case "static", "coingecko":
	default:
		return fmt.Errorf("unsupported CRYPTO_PRICE_PROVIDER %q", c.CryptoPriceProvider)
	}

	switch c.FXProvider {
	case "static", "frankfurter":
	default:
		return fmt.Errorf("unsupported FX_PROVIDER %q", c.FXProvider)
	}

	switch c.NewsProvider {
	case "static", "newsapiorg":
	default:
		return fmt.Errorf("unsupported FX_PROVIDER %q", c.FXProvider)
	}

	if c.EquityPriceProvider == "finnhub" && c.FinnhubAPIKey == "" {
		return fmt.Errorf("FINNHUB_API_KEY is required when PRICE_PROVIDER=finnhub")
	}

	if c.NewsProvider == "newsapiorg" && c.NewsAPIOrgAPIKey == "" {
		return fmt.Errorf("NEWSAPIORG_API_KEY is required when PRICE_PROVIDER=newsapiorg")
	}

	if c.LLMEnabled {
		switch c.LLMProvider {
		case "ollama":
			if c.LLMBaseURL == "" {
				return fmt.Errorf("LLM_BASE_URL is required when LLM is enabled")
			}
			if c.LLMModel == "" {
				return fmt.Errorf("LLM_MODEL is required when LLM is enabled")
			}
		default:
			return fmt.Errorf("unsupported LLM_PROVIDER %q", c.LLMProvider)
		}
	}

	return nil
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

func (c Config) String() string {
	type kv struct {
		Key   string
		Value string
	}

	rows := []kv{
		{"PRICE_PROVIDER", c.EquityPriceProvider},
		{"FX_PROVIDER", c.FXProvider},
		{"CACHE_ENABLED", fmt.Sprintf("%t", c.CacheEnabled)},
		{"CACHE_PRICE_TTL", c.PriceCacheTTL.String()},
		{"CACHE_FX_TTL", c.FXCacheTTL.String()},

		{"", ""}, // spacer

		{"LLM_ENABLED", fmt.Sprintf("%t", c.LLMEnabled)},
		{"LLM_PROVIDER", c.LLMProvider},
		{"LLM_BASE_URL", c.LLMBaseURL},
		{"LLM_MODEL", c.LLMModel},

		{"", ""}, // spacer

		{"FINNHUB_API_KEY", mask(c.FinnhubAPIKey)},
	}

	// find max key length
	maxKeyLen := 0
	for _, r := range rows {
		if len(r.Key) > maxKeyLen {
			maxKeyLen = len(r.Key)
		}
	}

	var b strings.Builder
	b.WriteString("Configuration:\n")

	for _, r := range rows {
		if r.Key == "" {
			b.WriteString("\n")
			continue
		}

		fmt.Fprintf(&b, "  %-*s : %s\n", maxKeyLen, r.Key, r.Value)
	}

	return b.String()
}

func mask(s string) string {
	if s == "" {
		return "<not set>"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}

// func LoadKeywordHints(path string) (map[string][]string, error) {
// 	b, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("read keyword hints: %w", err)
// 	}

// 	var hints map[string][]string
// 	if err := json.Unmarshal(b, &hints); err != nil {
// 		return nil, fmt.Errorf("unmarshal keyword hints: %w", err)
// 	}

// 	return hints, nil
// }
