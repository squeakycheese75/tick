package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type GetPortfolioSummaryUseCase struct {
	portfolioSvc PortfolioSvc
}

func NewGetPortfolioSummaryUseCase(portfolioSvc PortfolioSvc) *GetPortfolioSummaryUseCase {
	return &GetPortfolioSummaryUseCase{
		portfolioSvc: portfolioSvc,
	}
}

func (uc *GetPortfolioSummaryUseCase) Execute(ctx context.Context, in GetPortfolioSummaryUsecaseInput) (GetPortfolioSummaryUsecaseOutput, error) {
	result, err := uc.portfolioSvc.GetAnalysis(ctx, in.PortfolioName)
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
			QuotedPrice:        pos.QuotedPrice,
			ConvertedPrice:     pos.ConvertedPrice,
			Weight:             pos.Weight,
			MarketValueBase:    pos.MarketValueBase,
			FXRate:             pos.FXRate,
			Quantity:           pos.Quantity,
		})
	}

	return result
}
