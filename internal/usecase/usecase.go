package usecase

import (
	"context"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
	"github.com/squeakycheese75/tick/internal/instruments"
	"github.com/squeakycheese75/tick/internal/repository"
)

//go:generate mockgen -destination=./mocks/mock_interfaces.go -package=mocks . PortfolioRepository,InstrumentRepository,PositionRepository,InstrumentResolver

type (
	PortfolioRepository interface {
		GetByName(ctx context.Context, name string) (repository.Portfolio, error)
		Create(ctx context.Context, p repository.Portfolio) error
	}
	PositionRepository interface {
		ListByPortfolioID(ctx context.Context, portfolioID int64) ([]repository.Position, error)
		Create(ctx context.Context, p repository.CreatePositionParams) error
	}
	InstrumentRepository interface {
		Create(ctx context.Context, p repository.Instrument) (repository.Instrument, error)
		GetBySymbolAndExchange(ctx context.Context, symbol, exchange string) (repository.Instrument, error)
		GetOrCreate(ctx context.Context, in repository.Instrument) (repository.Instrument, error)
	}
	InstrumentResolver interface {
		Resolve(ctx context.Context, symbol string) (instruments.ResolvedInstrument, error)
	}
	PortfolioSnapshotRepository interface {
		Create(ctx context.Context, in repository.PortfolioSnapshot, positions []repository.PortfolioSnapshotPosition) (int64, error)
		GetLatestBefore(ctx context.Context, portfolioName string, before time.Time) (repository.PortfolioSnapshot, error)
		ListPositionsBySnapshotID(ctx context.Context, snapshotID int64) ([]repository.PortfolioSnapshotPosition, error)
	}
)

type (
	PriceProvider interface {
		GetPrice(ctx context.Context, ticker string) (price float64, currency string, err error)
	}
	FXProvider interface {
		GetRate(ctx context.Context, from string, to string) (float64, error)
	}
	PortfolioAnalyzer interface {
		Analyze(ctx context.Context, in analysis.AnalyzePortfolioInput) (domain.PortfolioAnalysis, error)
	}
	RiskAnalyzer interface {
		Analyze(in domain.PortfolioAnalysis) (analysis.PortfolioRisk, error)
	}
	NewsProvider interface {
		GetNews(ctx context.Context, ticker string, limit int) ([]domain.NewsHeadline, error)
	}
	PortfolioSvc interface {
		GetAnalysis(ctx context.Context, portfolioName string) (domain.PortfolioAnalysis, error)
		GetRisk(ctx context.Context, portfolioName string) (domain.PortfolioRisk, error)
	}
	PortfolioInsights interface {
		TopHoldings(portfolioAnalysis domain.PortfolioAnalysis, limit int) []domain.AnalyzedPosition
		AttentionSignals(portfolioAnalysis domain.PortfolioAnalysis, portfolioRisk analysis.PortfolioRisk) []string
	}
	DailyReportSummarizer interface {
		Summarize(ctx context.Context, report domain.DailyReport) (string, error)
		Enabled() bool
	}
	ReportBuilder interface {
		BuildDailyReport(ctx context.Context, portfolioName string, newsLimit int) (domain.DailyReportResult, error)
	}
)
