package app

import (
	"fmt"

	"github.com/squeakycheese75/tick/internal/adapters/market"
	"github.com/squeakycheese75/tick/internal/service"
)

func BuildPriceProvider(cfg Config) (service.PriceProvider, error) {
	var provider service.PriceProvider

	switch cfg.PriceProvider {
	case "static":
		provider = market.NewStaticPriceProvider()

	case "finnhub":
		provider = market.NewFinnhubPriceProvider(cfg.FinnhubAPIKey)

	default:
		return nil, fmt.Errorf("unsupported PRICE_PROVIDER %q", cfg.PriceProvider)
	}

	if cfg.CacheEnabled {
		provider = market.NewCachedPriceProvider(provider, cfg.PriceCacheTTL)
	}

	return provider, nil
}

func BuildFXProvider(cfg Config) (service.FXProvider, error) {
	var provider service.FXProvider

	switch cfg.FXProvider {
	case "static":
		provider = market.NewStaticFXProvider()

	case "frankfurter":
		provider = market.NewFrankfurterFXProvider()

	default:
		return nil, fmt.Errorf("unsupported FX_PROVIDER %q", cfg.FXProvider)
	}

	// ✅ apply cache conditionally
	if cfg.CacheEnabled {
		provider = market.NewCachedFXProvider(provider, cfg.FXCacheTTL)
	}

	return provider, nil
}
