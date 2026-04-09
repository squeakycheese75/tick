package app

import (
	"github.com/squeakycheese75/tick/internal/adapters/market"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
	"github.com/squeakycheese75/tick/internal/store"
	"github.com/squeakycheese75/tick/internal/usecase"
)

type Runtime struct {
	GetPortfolioSummary *usecase.GetPortfolioSummaryUseCase
	CreatePortfolio     *usecase.CreatePortfolioUseCase
	AddPosition         *usecase.AddPositionToPortfolioUseCase
	GetPortfolioRisk    *usecase.GetPortfolioRiskUseCase
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

	portfolioAnalyser := analysis.NewPortfolioAnalyzer(priceProvider, fxProvider)
	riskAnalyser := analysis.NewRiskAnalyzer()

	return &Runtime{
		GetPortfolioSummary: usecase.NewGetPortfolioSummaryUseCase(
			portfolioRepo,
			positionRepo,
			portfolioAnalyser,
		),
		CreatePortfolio: usecase.NewCreatePortfolioUseCase(portfolioRepo),
		AddPosition:     usecase.NewAddPositionToPortfolioUseCase(positionRepo, portfolioRepo),
		GetPortfolioRisk: usecase.NewGetPortfolioRiskUseCase(
			portfolioRepo,
			positionRepo,
			portfolioAnalyser,
			riskAnalyser,
		),
	}, nil
}
