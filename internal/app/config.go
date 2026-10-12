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

	// APIToken is the bearer token `tick serve` requires; the server will not start without one.
	APIToken string

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
		EquityPriceProviders:    splitList(firstEnv("static", "EQUITY_PRICE_PROVIDERS", "EQUITY_PRICE_PROVIDER")),
		CryptoPriceProviders:    splitList(firstEnv("static", "CRYPTO_PRICE_PROVIDERS", "CRYPTO_PRICE_PROVIDER")),
		CommodityPriceProviders: splitList(firstEnv("static", "COMMODITY_PRICE_PROVIDERS", "COMMODITY_PRICE_PROVIDER")),
		FXProviders:             splitList(firstEnv("static", "FX_PROVIDERS", "FX_PROVIDER")),
		NewsProviders:           splitList(firstEnv("static", "NEWS_PROVIDERS", "NEWS_PROVIDER")),
		// Funds are priced from prices you record yourself, which are always available,
		// so this works without any configuration. There is no "static" fund provider.
		FundPriceProviders: splitList(firstEnv("consumed", "FUND_PRICES_PROVIDERS", "FUND_PRICE_PROVIDERS", "FUND_PRICE_PROVIDER")),

		FinnhubAPIKey: os.Getenv("FINNHUB_API_KEY"),
		// NEWSAPIORG_API_KET is the misspelling `tick config init` used to write.
		NewsAPIOrgAPIKey: firstEnv("", "NEWSAPIORG_API_KEY", "NEWSAPIORG_API_KET"),
		APIToken:         os.Getenv("TICK_API_TOKEN"),
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
			return fmt.Errorf("unsupported EQUITY_PRICE_PROVIDERS %q", provider)
		}

		if provider == "finnhub" && c.FinnhubAPIKey == "" {
			return fmt.Errorf("FINNHUB_API_KEY is required when EQUITY_PRICE_PROVIDERS includes finnhub")
		}
	}

	for _, provider := range c.CryptoPriceProviders {
		switch provider {
		case "static", "coingecko", "yahoo":
		default:
			return fmt.Errorf("unsupported CRYPTO_PRICE_PROVIDERS %q", provider)
		}
	}

	for _, provider := range c.FXProviders {
		switch provider {
		case "static", "frankfurter":
		default:
			return fmt.Errorf("unsupported FX_PROVIDERS %q", provider)
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
		case "consumed":
		default:
			return fmt.Errorf("unsupported FUND_PRICES_PROVIDERS %q (the only fund provider is \"consumed\")", provider)
		}
	}

	for _, provider := range c.NewsProviders {
		switch provider {
		case "static", "newsapiorg":
		default:
			return fmt.Errorf("unsupported NEWS_PROVIDERS %q", provider)
		}

		if provider == "newsapiorg" && c.NewsAPIOrgAPIKey == "" {
			return fmt.Errorf("NEWSAPIORG_API_KEY is required when NEWS_PROVIDERS includes newsapiorg")
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

// firstEnv returns the first of keys that is set to something, or fallback. The plural names
// are the real ones; the singular ones are what older config files, and the template
// `tick config init` wrote, use, so those files keep working.
func firstEnv(fallback string, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}

	return fallback
}

// KnownConfigKeys lists every environment variable the configuration reads, aliases included.
func KnownConfigKeys() []string {
	return []string{
		"EQUITY_PRICE_PROVIDERS", "EQUITY_PRICE_PROVIDER",
		"CRYPTO_PRICE_PROVIDERS", "CRYPTO_PRICE_PROVIDER",
		"COMMODITY_PRICE_PROVIDERS", "COMMODITY_PRICE_PROVIDER",
		"FX_PROVIDERS", "FX_PROVIDER",
		"NEWS_PROVIDERS", "NEWS_PROVIDER",
		"FUND_PRICES_PROVIDERS", "FUND_PRICE_PROVIDERS", "FUND_PRICE_PROVIDER",
		"FINNHUB_API_KEY", "NEWSAPIORG_API_KEY", "NEWSAPIORG_API_KET", "TICK_API_TOKEN",
		"CACHE_ENABLED", "CACHE_PRICE_TTL", "CACHE_FX_TTL", "CONSUMED_PRICE_MAX_AGE",
		"LLM_ENABLED", "LLM_PROVIDER", "LLM_BASE_URL", "LLM_MODEL",
	}
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
		{"TICK_API_TOKEN", mask(c.APIToken)},
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

// splitList splits a comma separated list, dropping blanks.
func splitList(v string) []string {
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
