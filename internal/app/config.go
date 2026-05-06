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
	EquityPriceProviders    []string
	CryptoPriceProviders    []string
	CommodityPriceProviders []string
	FXProviders             []string
	NewsProviders           []string
	FundPriceProviders      []string

	FinnhubAPIKey    string
	NewsAPIOrgAPIKey string

	CacheEnabled  bool
	PriceCacheTTL time.Duration
	FXCacheTTL    time.Duration

	ConsumedPriceMaxAge time.Duration

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
		EquityPriceProviders:    splitEnvDefault("EQUITY_PRICE_PROVIDERS", "static"),
		CryptoPriceProviders:    splitEnvDefault("CRYPTO_PRICE_PROVIDERS", "static"),
		CommodityPriceProviders: splitEnvDefault("COMMODITY_PRICE_PROVIDERS", "static"),
		FXProviders:             splitEnvDefault("FX_PROVIDERS", "static"),
		NewsProviders:           splitEnvDefault("NEWS_PROVIDERS", "static"),
		FundPriceProviders:      splitEnvDefault("FUND_PRICES_PROVIDERS", "static"),

		FinnhubAPIKey:    os.Getenv("FINNHUB_API_KEY"),
		NewsAPIOrgAPIKey: os.Getenv("NEWSAPIORG_API_KEY"),
		CacheEnabled:     getenvDefault("CACHE_ENABLED", "true") == "true",

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

	cfg.ConsumedPriceMaxAge, err = durationEnv("CONSUMED_PRICE_MAX_AGE", 720*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("CONSUMED_PRICE_MAX_AGE: %w", err)
	}

	return cfg, nil
}

func (c Config) Validate() error {
	for _, provider := range c.EquityPriceProviders {
		switch provider {
		case "static", "finnhub", "yahoo":
		default:
			return fmt.Errorf("unsupported EQUITY_PRICE_PROVIDER %q", provider)
		}

		if provider == "finnhub" && c.FinnhubAPIKey == "" {
			return fmt.Errorf("FINNHUB_API_KEY is required when EQUITY_PRICE_PROVIDER includes finnhub")
		}
	}

	for _, provider := range c.CryptoPriceProviders {
		switch provider {
		case "static", "coingecko", "yahoo":
		default:
			return fmt.Errorf("unsupported CRYPTO_PRICE_PROVIDER %q", provider)
		}
	}

	for _, provider := range c.FXProviders {
		switch provider {
		case "static", "frankfurter":
		default:
			return fmt.Errorf("unsupported FX_PROVIDER %q", provider)
		}
	}

	for _, provider := range c.CommodityPriceProviders {
		switch provider {
		case "static", "yahoo":
		default:
			return fmt.Errorf("unsupported COMMODITY_PRICE_PROVIDERS %q", provider)
		}
	}

	for _, provider := range c.FundPriceProviders {
		switch provider {
		case "static", "consumed":
		default:
			return fmt.Errorf("unsupported FUND_PRICE_PROVIDERS %q", provider)
		}
	}

	for _, provider := range c.NewsProviders {
		switch provider {
		case "static", "newsapiorg":
		default:
			return fmt.Errorf("unsupported NEWS_PROVIDER %q", provider)
		}

		if provider == "newsapiorg" && c.NewsAPIOrgAPIKey == "" {
			return fmt.Errorf("NEWSAPIORG_API_KEY is required when NEWS_PROVIDER includes newsapiorg")
		}
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
		{"EQUITY_PRICE_PROVIDERS", strings.Join(c.EquityPriceProviders, ",")},
		{"FX_PROVIDERS", strings.Join(c.FXProviders, ",")},
		{"CRYPTO_PRICE_PROVIDERS", strings.Join(c.CryptoPriceProviders, ",")},
		{"COMMODITY_PRICE_PROVIDERS", strings.Join(c.CommodityPriceProviders, ",")},
		{"NEWS_PROVIDERS", strings.Join(c.NewsProviders, ",")},

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
		{"NEWSAPIORG_API_KEY", mask(c.NewsAPIOrgAPIKey)},
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

func splitEnvDefault(key, fallback string) []string {
	v := getenvDefault(key, fallback)

	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		out = append(out, p)
	}

	return out
}
