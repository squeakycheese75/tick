package app

import (
	"time"

	"github.com/squeakycheese75/tick/internal/adapters/market"
	"github.com/squeakycheese75/tick/internal/adapters/news"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
	"github.com/squeakycheese75/tick/internal/service"
	"github.com/squeakycheese75/tick/internal/store"
	"github.com/squeakycheese75/tick/internal/usecase"
)

type Runtime struct {
	GetPortfolioSummary *usecase.GetPortfolioSummaryUseCase
	CreatePortfolio     *usecase.CreatePortfolioUseCase
	AddPosition         *usecase.AddPositionToPortfolioUseCase
	GetPortfolioRisk    *usecase.GetPortfolioRiskUseCase
	GetDailyBrief       *usecase.GetDailyBriefUseCase
}

func BuildRuntime(dbPath string) (*Runtime, error) {
	// cfg, err := LoadConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	db, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}

	portfolioRepo := store.NewPortfolioRepository(db)
	positionRepo := store.NewPositionRepository(db)

	priceProvider := market.NewCachedPriceProvider(market.NewStaticPriceProvider(), 15*time.Minute)
	// priceProvider := market.NewFinnhubPriceProvider(cfg.FinnhubAPIKey)
	fxProvider := market.NewCachedFXProvider(market.NewStaticFXProvider(), 12*time.Hour)

	pricingSvc := service.NewPricingService(priceProvider, fxProvider)
	newsProvider := news.NewStaticProvider()

	portfolioAnalyser := analysis.NewPortfolioAnalyzer(pricingSvc)
	riskAnalyser := analysis.NewRiskAnalyzer()

	portfolioSvc := service.NewPortfolioService(portfolioRepo, positionRepo, portfolioAnalyser, riskAnalyser)
	portfolioInsights := service.NewPortfolioInsights()

	return &Runtime{
		GetPortfolioSummary: usecase.NewGetPortfolioSummaryUseCase(
			portfolioSvc,
		),
		CreatePortfolio: usecase.NewCreatePortfolioUseCase(portfolioRepo),
		AddPosition:     usecase.NewAddPositionToPortfolioUseCase(positionRepo, portfolioRepo),
		GetPortfolioRisk: usecase.NewGetPortfolioRiskUseCase(
			portfolioSvc,
		),
		GetDailyBrief: usecase.NewGetDailyBriefUseCase(portfolioSvc, portfolioInsights, newsProvider),
	}, nil
}
