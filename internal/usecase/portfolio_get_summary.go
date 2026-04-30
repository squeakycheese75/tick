package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type GetPortfolioSummaryUseCase struct {
	anaysisSvc AnaysisSvc
}

func NewGetPortfolioSummaryUseCase(anaysisSvc AnaysisSvc) *GetPortfolioSummaryUseCase {
	return &GetPortfolioSummaryUseCase{
		anaysisSvc: anaysisSvc,
	}
}

func (uc *GetPortfolioSummaryUseCase) Execute(ctx context.Context, in domain.GetPortfolioSummaryUsecaseInput) (domain.GetPortfolioSummaryUsecaseOutput, error) {
	result, err := uc.anaysisSvc.GetAnalysis(ctx, in.PortfolioName)
	if err != nil {
		return domain.GetPortfolioSummaryUsecaseOutput{}, fmt.Errorf("analyze portfolio: %w", err)
	}

	return mapPortfolioAnalysisToSummaryOutput(result), nil
}

func mapPortfolioAnalysisToSummaryOutput(r domain.PortfolioAnalysis) domain.GetPortfolioSummaryUsecaseOutput {
	result := domain.GetPortfolioSummaryUsecaseOutput{
		Report: domain.PortfoloSummaryReport{
			PortfolioName: r.PortfolioName,
			BaseCurrency:  r.BaseCurrency,
			TotalValue:    r.TotalValue,
			Positions:     make([]domain.SummaryPosition, 0, len(r.AnalyzedPositions)),
		},
	}

	var totalCost float64
	var totalPnL float64

	for _, pos := range r.AnalyzedPositions {
		costBasis := pos.Quantity * pos.AvgCost * pos.FXRate
		pnl := pos.MarketValueBase - costBasis

		var pnlPct float64
		if costBasis != 0 {
			pnlPct = pnl / costBasis
		}

		result.Report.Positions = append(result.Report.Positions, domain.SummaryPosition{
			Symbol:             pos.Symbol,
			BaseCurrency:       r.BaseCurrency,
			InstrumentCurrency: pos.InstrumentCurrency,
			Quantity:           pos.Quantity,
			QuotedPrice:        pos.QuotedPrice,
			AvgCost:            pos.AvgCost,
			FXRate:             pos.FXRate,
			MarketValueBase:    pos.MarketValueBase,
			CostBasisBase:      costBasis,
			UnrealizedPnL:      pnl,
			UnrealizedPnLPct:   pnlPct,
			Weight:             pos.Weight,
		})

		totalCost += costBasis
		totalPnL += pnl
	}

	var totalPnLPct float64
	if totalCost != 0 {
		totalPnLPct = totalPnL / totalCost
	}

	result.Report.TotalCost = totalCost
	result.Report.TotalPnL = totalPnL
	result.Report.TotalPnLPct = totalPnLPct

	return result
}
