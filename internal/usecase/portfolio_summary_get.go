package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type GetPortfolioSummaryUseCase struct {
	portfolios PortfolioRepository
	positions  PositionRepository
	analyzer   PortfolioAnalyzer
}

func NewGetPortfolioSummaryUseCase(portfolioRepo PortfolioRepository, positionRepo PositionRepository, analyzer PortfolioAnalyzer) *GetPortfolioSummaryUseCase {
	return &GetPortfolioSummaryUseCase{
		portfolios: portfolioRepo,
		positions:  positionRepo,
		analyzer:   analyzer,
	}
}

func (uc *GetPortfolioSummaryUseCase) Execute(ctx context.Context, in GetPortfolioSummaryUsecaseInput) (GetPortfolioSummaryUsecaseOutput, error) {
	pf, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf(
				"portfolio %q not found. Create it with:\n  tick portfolio create %s --base-currency EUR",
				in.PortfolioName,
				in.PortfolioName,
			)
		}
		return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("get portfolio: %w", err)
	}

	positions, err := uc.positions.ListByPortfolio(ctx, in.PortfolioName)
	if err != nil {
		return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("list positions: %w", err)
	}

	result, err := uc.analyzer.Analyze(ctx, analysis.AnalyzePortfolioInput{
		Portfolio: pf,
		Positions: positions,
	})
	if err != nil {
		return GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	return mapPortfolioAnalysisToSummaryOutput(result), nil
}

func mapPortfolioAnalysisToSummaryOutput(r analysis.PortfolioAnalysis) GetPortfolioSummaryUsecaseOutput {
	result := GetPortfolioSummaryUsecaseOutput{
		PortfolioName: r.PortfolioName,
		BaseCurrency:  r.BaseCurrency,
		TotalValue:    r.TotalValue,
		Positions:     make([]SummaryPosition, 0, len(r.AnalyzedPositions)),
	}

	for _, pos := range r.AnalyzedPositions {
		result.Positions = append(result.Positions, SummaryPosition{
			Ticker:             pos.Ticker,
			BaseCurrency:       r.BaseCurrency,
			InstrumentCurrency: pos.InstrumentCurrency,
			AvgCost:            pos.AvgCost,
			CurrentPrice:       pos.CurrentPrice,
			Weight:             pos.Weight,
			MarketValueBase:    pos.MarketValueBase,
			FXRate:             pos.FXRate,
		})
	}

	return result
}
