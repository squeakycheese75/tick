package market

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type NamedPriceProvider struct {
	Name     string
	Provider PriceProvider
}

type ChainPriceProvider struct {
	providers      []NamedPriceProvider
	symbolResolver SymbolResolver
}

func NewChainPriceProvider(
	providers []NamedPriceProvider,
	symbolResolver SymbolResolver,
) *ChainPriceProvider {
	return &ChainPriceProvider{
		providers:      providers,
		symbolResolver: symbolResolver,
	}
}

func (p *ChainPriceProvider) GetQuote(
	ctx context.Context,
	in GetQuoteParams,
) (domain.Quote, error) {
	var errs []error

	for _, candidate := range p.providers {
		if candidate.Provider == nil {
			continue
		}

		providerSymbol, err := p.resolveSymbol(in.Symbol, candidate.Name)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s resolve %s: %w", candidate.Name, in.Symbol, err))
			continue
		}

		quote, err := candidate.Provider.GetQuote(ctx, GetQuoteParams{
			Symbol:         in.Symbol,
			ProviderSymbol: providerSymbol,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("%s quote %s: %w", candidate.Name, providerSymbol, err))
			continue
		}

		quote.Symbol = in.Symbol
		quote.Source = candidate.Name

		return quote, nil
	}

	return domain.Quote{}, fmt.Errorf(
		"no price provider succeeded for %s: %w",
		in.Symbol,
		errors.Join(errs...),
	)
}

func (p *ChainPriceProvider) resolveSymbol(symbol, provider string) (string, error) {
	if p.symbolResolver == nil {
		return symbol, nil
	}

	resolvedSymbol, err := p.symbolResolver.Resolve(symbol, provider)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(resolvedSymbol) == "" {
		return "", fmt.Errorf("empty resolved symbol")
	}

	return resolvedSymbol, nil
}
