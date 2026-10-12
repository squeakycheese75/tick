package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/config"
)

// These tests change the environment and the working directory, so they cannot run in parallel.
// Each starts from nothing: no variables, an empty home, and a working directory with no .env.
// Running in the repository would pick up the developer's own .env and hide a bad default.

func isolate(t *testing.T) (home string) {
	t.Helper()

	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Chdir(t.TempDir())

	for _, key := range app.KnownConfigKeys() {
		old, had := os.LookupEnv(key)
		_ = os.Unsetenv(key)

		t.Cleanup(func() {
			if had {
				_ = os.Setenv(key, old)
				return
			}
			_ = os.Unsetenv(key)
		})
	}

	return home
}

func writeConfig(t *testing.T, home, content string) {
	t.Helper()

	dir := filepath.Join(home, ".tick")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.env"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestLoadConfig_NothingConfiguredIsValid(t *testing.T) {
	isolate(t)

	cfg, err := app.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("the defaults must be a valid configuration, got: %v", err)
	}

	if got := strings.Join(cfg.FundPriceProviders, ","); got != "consumed" {
		t.Fatalf("fund providers = %q, want consumed", got)
	}
	if got := strings.Join(cfg.EquityPriceProviders, ","); got != "static" {
		t.Fatalf("equity providers = %q, want static", got)
	}
}

func TestBuildRuntime_WorksWithNothingConfigured(t *testing.T) {
	home := isolate(t)

	rt, err := app.BuildRuntime(filepath.Join(home, "tick.db"))
	if err != nil {
		t.Fatalf("a fresh install must be able to start: %v", err)
	}

	if rt.ListPortfolios == nil || rt.ImportTrades == nil || rt.GetPortfolioSummary == nil || rt.GetDailyReport == nil {
		t.Fatalf("runtime is missing use cases: %+v", rt)
	}
}

func TestLoadConfig_OlderConfigFilesStillWork(t *testing.T) {
	home := isolate(t)

	// What `tick config init` used to write: singular names, and a misspelled news key.
	writeConfig(t, home, `
EQUITY_PRICE_PROVIDER=finnhub
CRYPTO_PRICE_PROVIDER=coingecko
FX_PROVIDER=frankfurter
NEWS_PROVIDER=newsapiorg
COMMODITY_PRICE_PROVIDER=yahoo
FINNHUB_API_KEY=fk
NEWSAPIORG_API_KET=nk
`)

	cfg, err := app.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	got := map[string]string{
		"equity":    strings.Join(cfg.EquityPriceProviders, ","),
		"crypto":    strings.Join(cfg.CryptoPriceProviders, ","),
		"fx":        strings.Join(cfg.FXProviders, ","),
		"news":      strings.Join(cfg.NewsProviders, ","),
		"commodity": strings.Join(cfg.CommodityPriceProviders, ","),
		"newsKey":   cfg.NewsAPIOrgAPIKey,
	}
	want := map[string]string{
		"equity": "finnhub", "crypto": "coingecko", "fx": "frankfurter", "news": "newsapiorg",
		"commodity": "yahoo", "newsKey": "nk",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q (old names must keep working, not fall back to static)", k, got[k], v)
		}
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("an older config should validate: %v", err)
	}
}

func TestLoadConfig_PluralNamesWinOverSingular(t *testing.T) {
	home := isolate(t)

	writeConfig(t, home, "EQUITY_PRICE_PROVIDER=yahoo\nEQUITY_PRICE_PROVIDERS=finnhub\nFINNHUB_API_KEY=k\nNEWSAPIORG_API_KEY=real\nNEWSAPIORG_API_KET=typo\n")

	cfg, err := app.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got := strings.Join(cfg.EquityPriceProviders, ","); got != "finnhub" {
		t.Fatalf("equity = %q, want finnhub", got)
	}
	if cfg.NewsAPIOrgAPIKey != "real" {
		t.Fatalf("news key = %q, want the correctly spelled one", cfg.NewsAPIOrgAPIKey)
	}
}

func TestLoadConfig_BadDurationIsReported(t *testing.T) {
	home := isolate(t)

	writeConfig(t, home, "CACHE_PRICE_TTL=soon\n")

	if _, err := app.LoadConfig(); err == nil || !strings.Contains(err.Error(), "CACHE_PRICE_TTL") {
		t.Fatalf("err = %v, want it to name the setting", err)
	}
}

// What `tick config init` writes must be a configuration tick actually reads. Every key in the
// template has to be one the loader knows (a misspelled key is silently ignored otherwise), and
// once the two API keys are filled in, the result must validate and mean what the template says.
func TestDefaultTemplateIsWhatTheLoaderReads(t *testing.T) {
	home := isolate(t)

	values, err := godotenv.Unmarshal(config.DefaultEnv)
	if err != nil {
		t.Fatalf("the template does not parse: %v", err)
	}

	known := map[string]bool{}
	for _, k := range app.KnownConfigKeys() {
		known[k] = true
	}
	for key := range values {
		if !known[key] {
			t.Errorf("the template sets %q, which tick does not read", key)
		}
	}

	writeConfig(t, home, config.DefaultEnv+"\nFINNHUB_API_KEY=fk\nNEWSAPIORG_API_KEY=nk\n")

	cfg, err := app.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("the template with its API keys filled in must validate: %v", err)
	}

	checks := map[string][]string{
		"EQUITY_PRICE_PROVIDERS":    cfg.EquityPriceProviders,
		"CRYPTO_PRICE_PROVIDERS":    cfg.CryptoPriceProviders,
		"COMMODITY_PRICE_PROVIDERS": cfg.CommodityPriceProviders,
		"FX_PROVIDERS":              cfg.FXProviders,
		"NEWS_PROVIDERS":            cfg.NewsProviders,
		"FUND_PRICES_PROVIDERS":     cfg.FundPriceProviders,
	}
	for key, got := range checks {
		if want := values[key]; strings.Join(got, ",") != want {
			t.Errorf("%s = %q, but the template says %q", key, strings.Join(got, ","), want)
		}
	}
}

func TestValidate_FundProviders(t *testing.T) {
	isolate(t)

	cfg, err := app.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	cfg.FundPriceProviders = []string{"static"}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "FUND_PRICES_PROVIDERS") {
		t.Fatalf("err = %v: there is no static fund provider, and the setting should be named", err)
	}

	if _, err := app.BuildFundPriceProvider(cfg, nil, nil); err == nil || !strings.Contains(err.Error(), "FUND_PRICES_PROVIDERS") {
		t.Fatalf("building a static fund provider should fail naming the setting, got %v", err)
	}

	cfg.FundPriceProviders = []string{"consumed"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("consumed must be valid: %v", err)
	}
	if _, err := app.BuildFundPriceProvider(cfg, nil, nil); err != nil {
		t.Fatalf("consumed must build: %v", err)
	}
}

// Every command builds the runtime, so a switched-on model that is not running must not stop
// `tick portfolio list`. Only asking for an AI summary needs it.
func TestBuildRuntime_DoesNotNeedTheModelUnlessAskedFor(t *testing.T) {
	home := isolate(t)

	// Nothing listens on port 1.
	writeConfig(t, home, "LLM_ENABLED=true\nLLM_PROVIDER=ollama\nLLM_BASE_URL=http://127.0.0.1:1\nLLM_MODEL=none\n")

	rt, err := app.BuildRuntime(filepath.Join(home, "tick.db"))
	if err != nil {
		t.Fatalf("a model that is not running must not stop the app starting: %v", err)
	}
	if rt.GetDailyReport == nil {
		t.Fatal("the daily report is still wired")
	}
}

func TestBuildLLMClient_RejectsAnUnknownProviderAndIgnoresADisabledOne(t *testing.T) {
	isolate(t)

	cfg, err := app.LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	cfg.LLMEnabled, cfg.LLMProvider = true, "mystery"
	if _, err := app.BuildLLMClient(cfg); err == nil || !strings.Contains(err.Error(), "LLM_PROVIDER") {
		t.Fatalf("err = %v, want an unsupported provider error", err)
	}

	cfg.LLMEnabled = false
	client, err := app.BuildLLMClient(cfg)
	if err != nil || client != nil {
		t.Fatalf("a disabled model builds nothing, got %v, %v", client, err)
	}
}
