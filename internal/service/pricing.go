package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
)

type (
	PriceProvider interface {
		GetQuote(ctx context.Context, ticker string) (domain.Quote, error)
	}
	FXProvider interface {
		GetRate(ctx context.Context, from string, to string) (domain.FXRate, error)
	}
)

type PricingService struct {
	priceProvider       PriceProvider
	cryptoPriceProvider PriceProvider
	fx                  FXProvider
}

func NewPricingService(equityPrices PriceProvider, cryptoPrices PriceProvider, fx FXProvider) *PricingService {
	return &PricingService{
		priceProvider:       equityPrices,
		cryptoPriceProvider: cryptoPrices,
		fx:                  fx,
	}
}

func (s *PricingService) GetValuationQuote(
	ctx context.Context,
	symbol string,
	targetCurrency string,
	instrumentCurrency string,
	instrumentType string,
) (domain.ValuationQuote, error) {
	quote, err := s.getQuote(ctx, symbol, instrumentType)
	if err != nil {
		return domain.ValuationQuote{}, err
	}

	priceCurrency := quote.PriceCurrency
	if priceCurrency == "" {
		priceCurrency = instrumentCurrency
	}
	if priceCurrency == "" {
		return domain.ValuationQuote{}, fmt.Errorf("missing price currency for %s", symbol)
	}

	pc := strings.ToUpper(priceCurrency)
	tc := strings.ToUpper(targetCurrency)

	quote.PriceCurrency = pc

	if pc == tc {
		return domain.ValuationQuote{
			Quote:                  quote,
			TargetCurrency:         tc,
			FXRate:                 1.0,
			ConvertedPrice:         quote.Price,
			ConvertedPreviousClose: quote.PreviousClose,
			ConvertedChange:        quote.Change,
		}, nil
	}

	rate, err := s.fx.GetRate(ctx, pc, tc)
	if err != nil {
		return domain.ValuationQuote{}, err
	}

	return domain.ValuationQuote{
		Quote:                  quote,
		TargetCurrency:         tc,
		FXRate:                 rate.Rate,
		ConvertedPrice:         quote.Price * rate.Rate,
		ConvertedPreviousClose: quote.PreviousClose * rate.Rate,
		ConvertedChange:        quote.Change * rate.Rate,
	}, nil
}

func (s *PricingService) getQuote(ctx context.Context, symbol, instrumentType string) (domain.Quote, error) {
	switch instrumentType {
	case string(domain.InstrumentTypeCrypto):
		return s.cryptoPriceProvider.GetQuote(ctx, symbol)
	case string(domain.InstrumentTypeEquity):
		return s.priceProvider.GetQuote(ctx, symbol)
	default:
		return domain.Quote{}, fmt.Errorf("unsupported ASSET_TYPE %q", instrumentType)
	}
}
