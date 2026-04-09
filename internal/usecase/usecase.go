package usecase

import (
	"context"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type (
	PortfolioRepository interface {
		GetByName(ctx context.Context, name string) (domain.Portfolio, error)
		Create(ctx context.Context, p domain.Portfolio) error
	}
	PositionRepository interface {
		ListByPortfolio(ctx context.Context, portfolioName string) ([]domain.Position, error)
		Create(ctx context.Context, p domain.Position) error
	}
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
