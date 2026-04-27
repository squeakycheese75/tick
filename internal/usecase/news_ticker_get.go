package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type GetTickerNewsUseCase struct {
	newsSvc NewsSvc
}

func NewGetTickerNewsUseCase(newsSvc NewsSvc) *GetTickerNewsUseCase {
	return &GetTickerNewsUseCase{newsSvc: newsSvc}
}

func (uc *GetTickerNewsUseCase) Execute(
	ctx context.Context,
	ticker string,
	limit int,
) (domain.NewsSummary, error) {

	if ticker == "" {
		return domain.NewsSummary{}, fmt.Errorf("ticker required")
	}

	if limit <= 0 {
		limit = 5
	}

	return uc.newsSvc.GetNews(ctx, ticker, limit)
}
