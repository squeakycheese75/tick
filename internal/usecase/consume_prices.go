package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type ConsumePriceUsecase struct {
	repo ConsumedPriceRepository
}

func NewConsumePriceUsecase(repo ConsumedPriceRepository) *ConsumePriceUsecase {
	return &ConsumePriceUsecase{
		repo: repo,
	}
}

func (u *ConsumePriceUsecase) Execute(
	ctx context.Context,
	in domain.ConsumePriceUsecaseInput,
) error {
	symbol := strings.ToUpper(strings.TrimSpace(in.Symbol))
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	source := strings.TrimSpace(in.Source)

	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if currency == "" {
		return fmt.Errorf("currency is required")
	}
	if source == "" {
		source = "manual"
	}
	if in.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if in.AsOf.IsZero() {
		in.AsOf = time.Now()
	}

	return u.repo.Create(ctx, repository.ConsumedPrice{
		Symbol:   symbol,
		Price:    in.Price,
		Currency: currency,
		AsOf:     in.AsOf,
		Source:   source,
	})
}
