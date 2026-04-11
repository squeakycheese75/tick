package app

import (
	"github.com/squeakycheese75/tick/internal/adapters/news"
	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/service"
	"github.com/squeakycheese75/tick/internal/usecase"
)

type Runtime struct {
	GetPortfolioSummary *usecase.GetPortfolioSummaryUseCase
	CreatePortfolio     *usecase.CreatePortfolioUseCase
	AddPosition         *usecase.AddPositionToPortfolioUseCase
	GetPortfolioRisk    *usecase.GetPortfolioRiskUseCase
	GetDailyReport      *usecase.GetDailyReportUseCase
}

func BuildRuntime(dbPath string) (*Runtime, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	database, err := db.OpenAndMigrateSqlite(dbPath)
	if err != nil {
		return nil, err
	}

	portfolioRepo := repository.NewPortfolioRepository(database)
	positionRepo := repository.NewPositionRepository(database)
	instrumentRepo := repository.NewInstrumentRepository(database)

	priceProvider, err := BuildPriceProvider(cfg)
	if err != nil {
		return nil, err
	}

	fxProvider, err := BuildFXProvider(cfg)
	if err != nil {
		return nil, err
	}

	llmProvider, err := BuildLLMClient(cfg)
	if err != nil {
		return nil, err
	}

	pricingSvc := service.NewPricingService(priceProvider, fxProvider)
	newsProvider := news.NewStaticProvider()

	portfolioAnalyser := analysis.NewPortfolioAnalyzer(pricingSvc)
	riskAnalyser := analysis.NewRiskAnalyzer()

	portfolioSvc := service.NewPortfolioService(portfolioRepo, positionRepo, portfolioAnalyser, riskAnalyser)
	portfolioInsights := service.NewPortfolioInsights()
	newsSvc := service.NewNewsService(newsProvider)

	reportingSvc := service.NewReportService(portfolioSvc, newsSvc, portfolioInsights)

	return &Runtime{
		GetPortfolioSummary: usecase.NewGetPortfolioSummaryUseCase(
			portfolioSvc,
		),
		CreatePortfolio: usecase.NewCreatePortfolioUseCase(portfolioRepo),
		AddPosition:     usecase.NewAddPositionToPortfolioUseCase(positionRepo, portfolioRepo, instrumentRepo),
		GetPortfolioRisk: usecase.NewGetPortfolioRiskUseCase(
			portfolioSvc,
		),
		GetDailyReport: usecase.NewGetDailyReportUseCase(reportingSvc, llmProvider),
	}, nil
}
