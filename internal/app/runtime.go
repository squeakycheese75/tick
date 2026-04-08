package app

import (
	"github.com/squeakycheese75/tick/cmd/usecase"
	"github.com/squeakycheese75/tick/internal/adapters/market"
	"github.com/squeakycheese75/tick/internal/store"
)

type Runtime struct {
	GetPortfolioSummary *usecase.GetPortfolioSummaryUseCase
	CreatePortfolio     *usecase.CreatePortfolioUseCase
	AddPosition         *usecase.AddPositionToPortfolioUseCase
}

func BuildRuntime(dbPath string) (*Runtime, error) {
	db, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}

	portfolioRepo := store.NewPortfolioRepository(db)
	positionRepo := store.NewPositionRepository(db)
	priceProvider := market.NewStaticPriceProvider()
	fxProvider := market.NewStaticFXProvider()

	return &Runtime{
		GetPortfolioSummary: usecase.NewGetPortfolioSummaryUseCase(
			portfolioRepo,
			positionRepo,
			priceProvider,
			fxProvider,
		),
		CreatePortfolio: usecase.NewCreatePortfolioUseCase(portfolioRepo),
		AddPosition:     usecase.NewAddPositionToPortfolioUseCase(positionRepo, portfolioRepo),
	}, nil
}
