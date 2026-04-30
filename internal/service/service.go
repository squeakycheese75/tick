package service

import (
	"context"

	"github.com/squeakycheese75/tick/internal/analysis"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type (
	PortfolioRepository interface {
		GetByName(ctx context.Context, name string) (repository.Portfolio, error)
	}
	PositionRepository interface {
		ListByPortfolioID(ctx context.Context, portfolioID int64) ([]repository.Position, error)
	}
	TargetRepository interface {
		ListByPortfolio(ctx context.Context, portfolioID int64) ([]domain.Target, error)
	}
)

type (
	PortfolioAnalyzer interface {
		Analyze(ctx context.Context, in analysis.AnalyzePortfolioInput) (domain.PortfolioAnalysis, error)
	}
	RiskAnalyzer interface {
		Analyze(in domain.PortfolioAnalysis) (analysis.PortfolioRisk, error)
	}
)
