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
		GetRate(ctx context.Context, from string, to string) (float64, error)
	}
)

type PricingService struct {
	prices PriceProvider
	fx     FXProvider
}

func NewPricingService(prices PriceProvider, fx FXProvider) *PricingService {
	return &PricingService{
		prices: prices,
		fx:     fx,
	}
}

func (s *PricingService) GetValuationQuote(
	ctx context.Context,
	ticker string,
	targetCurrency string,
	instrumentCurrency string,
) (domain.ValuationQuote, error) {

	quote, err := s.prices.GetQuote(ctx, ticker)
	if err != nil {
		return domain.ValuationQuote{}, err
	}

	priceCurrency := quote.PriceCurrency
	if priceCurrency == "" {
		priceCurrency = instrumentCurrency
	}
	if priceCurrency == "" {
		return domain.ValuationQuote{}, fmt.Errorf("missing price currency for %s", ticker)
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
		FXRate:                 rate,
		ConvertedPrice:         quote.Price * rate,
		ConvertedPreviousClose: quote.PreviousClose * rate,
		ConvertedChange:        quote.Change * rate,
	}, nil
}
