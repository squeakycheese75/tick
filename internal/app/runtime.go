package app

import (
	"github.com/squeakycheese75/tick/internal/analysis"
	"github.com/squeakycheese75/tick/internal/db"
	"github.com/squeakycheese75/tick/internal/instruments"
	"github.com/squeakycheese75/tick/internal/report"
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
	ImportPortfolio     *usecase.ImportPortfolioUseCase
	GetTickerNews       *usecase.GetTickerNewsUseCase
	GetMorningBrief     *usecase.GetMorningBriefUsecase
	SetTarget           *usecase.SetTargetUseCase
	ListTargets         *usecase.ListTargetsUseCase
	RemoveTarget        *usecase.RemoveTargetUsecase
}

func BuildRuntime(dbPath string) (*Runtime, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	database, err := db.OpenAndMigrateSqlite(dbPath)
	if err != nil {
		return nil, err
	}

	portfolioRepo := repository.NewPortfolioRepository(database)
	positionRepo := repository.NewPositionRepository(database)
	instrumentRepo := repository.NewInstrumentRepository(database)
	snapshotRepo := repository.NewSnapshotRepository(database)
	targetRespository := repository.NewTargetRepository(database)

	// Caching
	priceCacher := repository.NewPriceCacheRepository(database)
	fxCacher := repository.NewFXCacheRepository(database)

	// Adapters/Providers
	equityPriceProvider, err := BuildEquityPriceProvider(cfg, priceCacher)
	if err != nil {
		return nil, err
	}

	cryptoPriceProvider, err := BuildCryptoPriceProvider(cfg, priceCacher)
	if err != nil {
		return nil, err
	}

	fxProvider, err := BuildFXProvider(cfg, fxCacher)
	if err != nil {
		return nil, err
	}

	newsProvider, err := BuildNewsProvider(cfg)
	if err != nil {
		return nil, err
	}

	llmProvider, err := BuildLLMClient(cfg)
	if err != nil {
		return nil, err
	}

	// Services
	pricingSvc := service.NewPricingService(equityPriceProvider, cryptoPriceProvider, fxProvider)

	portfolioAnalyser := analysis.NewPortfolioAnalyzer(pricingSvc)
	riskAnalyser := analysis.NewRiskAnalyzer()
	analysisSvc := service.NewPortfolioAnalysisSvc(portfolioRepo, positionRepo, portfolioAnalyser)
	riskSvc := service.NewPortfolioRiskSvc(portfolioRepo, positionRepo, riskAnalyser)

	portfolioInsights := service.NewInsightsSvc()
	newsSvc := service.NewNewsService(newsProvider)
	snapshotSvc := service.NewSnapshotService(snapshotRepo)
	targetSvc := service.NewTargetSvc(portfolioRepo, targetRespository)

	instrumentResolver, err := instruments.NewStaticResolver()
	if err != nil {
		return nil, err
	}

	reportingBuilder := report.NewReportBuilder(analysisSvc, riskSvc, pricingSvc, newsSvc, portfolioInsights, snapshotSvc, targetSvc)

	var summarizer usecase.DailyReportSummarizer = service.NoopSummarizer{}

	if llmProvider != nil {
		summarizer = service.NewLLMDailyReportSummarizer(llmProvider)
	}

	return &Runtime{
		GetPortfolioSummary: usecase.NewGetPortfolioSummaryUseCase(
			analysisSvc,
		),
		CreatePortfolio: usecase.NewCreatePortfolioUseCase(portfolioRepo),
		AddPosition:     usecase.NewAddPositionToPortfolioUseCase(positionRepo, portfolioRepo, instrumentRepo, instrumentResolver),
		GetPortfolioRisk: usecase.NewGetPortfolioRiskUseCase(
			analysisSvc,
			riskSvc,
		),
		GetDailyReport:  usecase.NewGetDailyReportUseCase(reportingBuilder, summarizer, snapshotRepo),
		ImportPortfolio: usecase.NewImportPortfolioUseCase(positionRepo, portfolioRepo, instrumentRepo),
		GetTickerNews:   usecase.NewGetTickerNewsUseCase(newsSvc),
		GetMorningBrief: usecase.NewGetMorningBriefUsecase(reportingBuilder),
		SetTarget:       usecase.NewSetTargetUseCase(portfolioRepo, targetRespository),
		ListTargets:     usecase.NewListTargetsUseCase(portfolioRepo, targetRespository),
		RemoveTarget:    usecase.NewRemoveTargetUsecase(portfolioRepo, targetRespository),
	}, nil
}
