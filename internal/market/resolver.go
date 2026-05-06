package market

import (
	"github.com/squeakycheese75/tick/internal/domain"
)

type StaticSymbolResolver struct {
	symbols map[string][]domain.ProviderSymbol
}

func NewStaticSymbolResolver(
	symbols map[string][]domain.ProviderSymbol,
) *StaticSymbolResolver {
	return &StaticSymbolResolver{
		symbols: symbols,
	}
}

func (r *StaticSymbolResolver) Resolve(symbol, provider string) (string, error) {
	providerSymbols, ok := r.symbols[symbol]
	if !ok {
		return symbol, nil
	}

	for _, s := range providerSymbols {
		if s.Provider == provider {
			return s.Symbol, nil
		}
	}

	return symbol, nil
}
