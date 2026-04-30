package analysis

import (
	"context"

	"github.com/squeakycheese75/tick/internal/domain"
)

//go:generate mockgen -destination=./mocks/mock_interfaces.go -package=mocks . PricingSvc

type (
	PricingSvc interface {
		GetValuationQuote(ctx context.Context, ticker string, targetCurrency string, instrumentCurrency string, instrumentType string) (domain.ValuationQuote, error)
	}
)
