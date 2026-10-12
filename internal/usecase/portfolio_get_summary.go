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
			PricesAsOf:    r.PricesAsOf(),
			Positions:     make([]domain.SummaryPosition, 0, len(r.AnalyzedPositions)),
		},
	}

	var totalCost, totalPnL, totalAssetPnL, totalFXPnL float64

	for _, pos := range r.AnalyzedPositions {
		sp := summarisePosition(r.BaseCurrency, pos)

		result.Report.Positions = append(result.Report.Positions, sp)

		totalCost += sp.CostBasisBase
		totalPnL += sp.UnrealizedPnL

		if sp.FXKnown {
			totalAssetPnL += sp.AssetPnL
			totalFXPnL += sp.FXPnL
		}
	}

	var totalPnLPct float64
	if totalCost != 0 {
		totalPnLPct = totalPnL / totalCost
	}

	result.Report.TotalCost = totalCost
	result.Report.TotalPnL = totalPnL
	result.Report.TotalPnLPct = totalPnLPct
	result.Report.TotalAssetPnL = totalAssetPnL
	result.Report.TotalFXPnL = totalFXPnL

	return result
}

// summarisePosition values one position and splits its PnL into the part from the asset price
// and the part from the FX rate moving since purchase:
//
//	asset PnL = market value - cost at today's FX rate
//	FX PnL    = cost at today's FX rate - cost at the purchase FX rate
//
// Without a recorded base cost the purchase rate is unknown, so cost is taken at today's rate
// and the whole PnL is reported as asset PnL.
func summarisePosition(baseCurrency string, pos domain.AnalyzedPosition) domain.SummaryPosition {
	costAtToday := pos.Quantity * pos.AvgCost * pos.FXRate
	fxKnown := pos.AvgCostBase > 0

	costBasis := costAtToday
	if fxKnown {
		costBasis = pos.Quantity * pos.AvgCostBase
	}

	pnl := pos.MarketValueBase - costBasis

	var pnlPct float64
	if costBasis != 0 {
		pnlPct = pnl / costBasis
	}

	var fxPnL float64
	if fxKnown {
		fxPnL = costAtToday - costBasis
	}

	return domain.SummaryPosition{
		Symbol:             pos.Symbol,
		BaseCurrency:       baseCurrency,
		InstrumentCurrency: pos.InstrumentCurrency,
		Quantity:           pos.Quantity,
		QuotedPrice:        pos.QuotedPrice,
		AvgCost:            pos.AvgCost,
		FXRate:             pos.FXRate,
		MarketValueBase:    pos.MarketValueBase,
		CostBasisBase:      costBasis,
		UnrealizedPnL:      pnl,
		AssetPnL:           pnl - fxPnL,
		FXPnL:              fxPnL,
		FXKnown:            fxKnown,
		UnrealizedPnLPct:   pnlPct,
		Weight:             pos.Weight,
	}
}
