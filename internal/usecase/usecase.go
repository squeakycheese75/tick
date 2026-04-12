package usecase

import (
	"context"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
	"github.com/squeakycheese75/tick/internal/repository"
)

//go:generate mockgen -destination=./mocks/mock_interfaces.go -package=mocks . PortfolioRepository,InstrumentRepository,PositionRepository

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
)

type (
	PriceProvider interface {
		GetPrice(ctx context.Context, ticker string) (price float64, currency string, err error)
	}
	FXProvider interface {
		GetRate(ctx context.Context, from string, to string) (float64, error)
	}
	PortfolioAnalyzer interface {
		Analyze(ctx context.Context, in analysis.AnalyzePortfolioInput) (analysis.PortfolioAnalysis, error)
	}
	RiskAnalyzer interface {
		Analyze(in analysis.PortfolioAnalysis) (analysis.PortfolioRisk, error)
	}
	NewsProvider interface {
		GetNews(ctx context.Context, ticker string, limit int) ([]domain.NewsHeadline, error)
	}
	PortfolioSvc interface {
		GetAnalysis(ctx context.Context, portfolioName string) (analysis.PortfolioAnalysis, error)
		GetRisk(ctx context.Context, portfolioName string) (analysis.PortfolioRisk, error)
	}
	PortfolioInsights interface {
		TopHoldings(portfolioAnalysis analysis.PortfolioAnalysis, limit int) []analysis.AnalyzedPosition
		AttentionSignals(portfolioAnalysis analysis.PortfolioAnalysis, portfolioRisk analysis.PortfolioRisk) []string
	}
)
