package app

import (
	"context"
	"fmt"
	"time"

	"github.com/squeakycheese75/tick/data"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/llm"
	"github.com/squeakycheese75/tick/internal/market"
	"github.com/squeakycheese75/tick/internal/news"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/service"
)

type (
	PriceProvider interface {
		GetQuote(ctx context.Context, p market.GetQuoteParams) (domain.Quote, error)
	}
	FXProvider interface {
		GetRate(ctx context.Context, from string, to string) (domain.FXRate, error)
	}
	PriceCacheStore interface {
		Upsert(ctx context.Context, quote repository.PriceQuote, fetchedAt time.Time) error
		Get(ctx context.Context, ticker string) (repository.CachedPriceQuote, error)
	}
	FXCacheStore interface {
		Get(ctx context.Context, baseCurrency, quoteCurrency string) (repository.CachedFXRate, error)
		Upsert(ctx context.Context, cached repository.CachedFXRate) error
	}
	ConsumedPriceStore interface {
		Create(ctx context.Context, price repository.ConsumedPrice) error
		GetLatest(ctx context.Context, symbol string) (repository.ConsumedPrice, error)
	}

	NewsProvider interface {
		GetNews(ctx context.Context, symbol string, limit int) (domain.NewsSummary, error)
	}
	LLMProvider interface {
		Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error)
		Ping(ctx context.Context) error
	}
	SymbolResolver interface {
		Resolve(symbol, provider string) (string, error)
	}
)

func BuildEquityPriceProvider(
	cfg Config,
	priceCacheStore PriceCacheStore,
	symbolResolver SymbolResolver,
) (PriceProvider, error) {
	providers := make([]market.NamedPriceProvider, 0)

	for _, name := range cfg.EquityPriceProviders {
		provider, err := buildSingleEquityPriceProvider(cfg, name, priceCacheStore)
		if err != nil {
			return nil, err
		}

		providers = append(providers, market.NamedPriceProvider{
			Name:     name,
			Provider: provider,
		})
	}

	return market.NewChainPriceProvider(providers, symbolResolver), nil
}

func buildSingleEquityPriceProvider(
	cfg Config,
	name string,
	priceCacheStore PriceCacheStore,
) (PriceProvider, error) {
	var provider service.PriceProvider

	switch name {
	case "static":
		provider = market.NewStaticPriceProvider()

	case "finnhub":
		provider = market.NewFinnhubPriceProvider(cfg.FinnhubAPIKey)

	case "yahoo":
		provider = market.NewYahooPriceProvider(nil)

	default:
		return nil, fmt.Errorf("unsupported EQUITY_PRICE_PROVIDER %q", name)
	}

	if cfg.CacheEnabled && priceCacheStore != nil {
		provider = market.NewCachedPriceProvider(provider, priceCacheStore, cfg.PriceCacheTTL)
	}

	return provider, nil
}

func BuildCryptoPriceProvider(
	cfg Config,
	priceCacheStore PriceCacheStore,
	symbolResolver SymbolResolver,
) (PriceProvider, error) {
	providers := make([]market.NamedPriceProvider, 0, len(cfg.CryptoPriceProviders))

	for _, name := range cfg.CryptoPriceProviders {
		provider, err := buildSingleCryptoPriceProvider(cfg, name, priceCacheStore)
		if err != nil {
			return nil, err
		}

		providers = append(providers, market.NamedPriceProvider{
			Name:     name,
			Provider: provider,
		})
	}

	return market.NewChainPriceProvider(providers, symbolResolver), nil
}

func buildSingleCryptoPriceProvider(
	cfg Config,
	name string,
	priceCacheStore PriceCacheStore,
) (PriceProvider, error) {
	var provider PriceProvider

	switch name {
	case "static":
		provider = market.NewStaticCryptoPriceProvider()

	case "coingecko":
		provider = market.NewCoinGeckoProvider()

	case "yahoo":
		provider = market.NewYahooPriceProvider(nil)

	default:
		return nil, fmt.Errorf("unsupported CRYPTO_PRICE_PROVIDER %q", name)
	}

	if cfg.CacheEnabled && priceCacheStore != nil {
		provider = market.NewCachedPriceProvider(provider, priceCacheStore, cfg.PriceCacheTTL)
	}

	return provider, nil
}

func BuildFXProvider(cfg Config, fxCacheStore FXCacheStore) (FXProvider, error) {
	if len(cfg.FXProviders) == 0 {
		return nil, fmt.Errorf("no FX providers configured")
	}

	name := cfg.FXProviders[0]

	var provider FXProvider

	switch name {
	case "static":
		provider = market.NewStaticFXProvider()

	case "frankfurter":
		provider = market.NewFrankfurterFXProvider()

	default:
		return nil, fmt.Errorf("unsupported FX_PROVIDER %q", name)
	}

	if cfg.CacheEnabled && fxCacheStore != nil {
		provider = market.NewCachedFXProvider(provider, fxCacheStore, cfg.FXCacheTTL)
	}

	return provider, nil
}

func BuildLLMClient(cfg Config) (LLMProvider, error) {
	var llmClient LLMProvider

	if cfg.LLMEnabled {
		switch cfg.LLMProvider {
		case "ollama":
			llmClient = llm.NewOllamaClient(cfg.LLMBaseURL, cfg.LLMModel)
		default:
			return nil, fmt.Errorf("unsupported LLM_PROVIDER %q", cfg.LLMProvider)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := llmClient.Ping(ctx); err != nil {
			return nil, fmt.Errorf("llm not ready: %w", err)
		}
	}

	return llmClient, nil
}

func BuildNewsProvider(cfg Config) (NewsProvider, error) {
	if len(cfg.NewsProviders) == 0 {
		return nil, fmt.Errorf("no news providers configured")
	}

	name := cfg.NewsProviders[0]

	var provider NewsProvider

	switch name {
	case "static":
		provider = news.NewStaticProvider()

	case "newsapiorg":
		keywordHints, err := data.LoadKeywordHints()
		if err != nil {
			return nil, err
		}

		provider = news.NewNewsAPIProvider(cfg.NewsAPIOrgAPIKey, keywordHints)

	default:
		return nil, fmt.Errorf("unsupported NEWS_PROVIDER %q", name)
	}

	return provider, nil
}

func BuildCommodityPriceProvider(
	cfg Config,
	priceCacheStore PriceCacheStore,
	symbolResolver SymbolResolver,
) (PriceProvider, error) {
	providers := make([]market.NamedPriceProvider, 0, len(cfg.CommodityPriceProviders))

	for _, name := range cfg.CommodityPriceProviders {
		provider, err := buildSingleCommodityPriceProvider(cfg, name, priceCacheStore)
		if err != nil {
			return nil, err
		}

		providers = append(providers, market.NamedPriceProvider{
			Name:     name,
			Provider: provider,
		})
	}

	return market.NewChainPriceProvider(providers, symbolResolver), nil
}

func buildSingleCommodityPriceProvider(
	cfg Config,
	name string,
	priceCacheStore PriceCacheStore,
) (PriceProvider, error) {
	var provider PriceProvider

	switch name {
	case "static":
		provider = market.NewStaticCommodityPriceProvider()

	case "yahoo":
		provider = market.NewYahooPriceProvider(nil)

	default:
		return nil, fmt.Errorf("unsupported COMMODITY_PRICE_PROVIDER %q", name)
	}

	if cfg.CacheEnabled && priceCacheStore != nil {
		provider = market.NewCachedPriceProvider(provider, priceCacheStore, cfg.PriceCacheTTL)
	}

	return provider, nil
}

func BuildFundPriceProvider(
	cfg Config,
	consumedPriceStore ConsumedPriceStore,
	symbolResolver SymbolResolver,
) (PriceProvider, error) {
	providers := make([]market.NamedPriceProvider, 0, len(cfg.FundPriceProviders))

	for _, name := range cfg.FundPriceProviders {
		provider, err := buildSingleFundPriceProvider(cfg, name, consumedPriceStore)
		if err != nil {
			return nil, err
		}

		providers = append(providers, market.NamedPriceProvider{
			Name:     name,
			Provider: provider,
		})
	}

	return market.NewChainPriceProvider(providers, symbolResolver), nil
}

func buildSingleFundPriceProvider(
	cfg Config,
	name string,
	consumedPriceStore ConsumedPriceStore,
) (PriceProvider, error) {
	switch name {
	case "consumed":
		return market.NewConsumedPriceProvider(
			consumedPriceStore,
			cfg.ConsumedPriceMaxAge,
		), nil

	// case "static":
	// 	return market.NewStaticFundPriceProvider(), nil

	default:
		return nil, fmt.Errorf("unsupported FUND_PRICE_PROVIDER %q", name)
	}
}
