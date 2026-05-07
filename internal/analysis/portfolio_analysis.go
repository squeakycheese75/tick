package analysis

import (
	"context"
	"errors"
	"sort"

	"github.com/squeakycheese75/tick/internal/domain"
)

type PortfolioAnalyzer struct {
	pricingSvc PricingSvc
}

func NewPortfolioAnalyzer(pricingSvc PricingSvc) *PortfolioAnalyzer {
	return &PortfolioAnalyzer{
		pricingSvc: pricingSvc,
	}
}

type AnalyzePortfolioInput struct {
	Portfolio domain.Portfolio
	Positions []domain.Position
}

func (a *PortfolioAnalyzer) Analyze(ctx context.Context, in AnalyzePortfolioInput) (domain.PortfolioAnalysis, error) {
	result := domain.PortfolioAnalysis{
		PortfolioName:     in.Portfolio.Name,
		BaseCurrency:      in.Portfolio.BaseCurrency,
		AnalyzedPositions: make([]domain.AnalyzedPosition, 0, len(in.Positions)),
		ValuationIssues:   make([]domain.ValuationIssue, 0),
	}

	for _, pos := range in.Positions {
		valuationQuote, err := a.pricingSvc.GetValuationQuote(ctx, pos.Instrument.Symbol, pos.Instrument.ProviderSymbol, in.Portfolio.BaseCurrency, pos.Instrument.QuoteCurrency, string(pos.Instrument.InstrumentType))
		if err != nil {
			result.ValuationIssues = append(result.ValuationIssues, domain.ValuationIssue{
				Symbol:         pos.Instrument.Symbol,
				InstrumentType: string(pos.Instrument.InstrumentType),
				Quantity:       pos.Quantity,
				Type:           classifyValuationIssue(err),
				Message:        cleanValuationMessage(err),
				Hint:           valuationHint(pos),
			})
			continue
		}

		marketValueBase := pos.Quantity * valuationQuote.ConvertedPrice
		result.TotalValue += marketValueBase

		result.AnalyzedPositions = append(result.AnalyzedPositions, domain.AnalyzedPosition{
			Symbol:             pos.Instrument.Symbol,
			Quantity:           pos.Quantity,
			AvgCost:            pos.AvgCost,
			InstrumentCurrency: pos.Instrument.QuoteCurrency,
			QuotedPrice:        valuationQuote.Quote.Price,
			QuotedChange:       valuationQuote.Quote.Change,
			QuotedChangePct:    valuationQuote.Quote.ChangePercent,
			PriceCurrency:      valuationQuote.Quote.PriceCurrency,
			FXRate:             valuationQuote.FXRate,
			ConvertedPrice:     valuationQuote.ConvertedPrice,
			ConvertedChange:    valuationQuote.Quote.Change * valuationQuote.FXRate,
			MarketValueBase:    marketValueBase,
		})
	}

	if result.TotalValue > 0 {
		for i := range result.AnalyzedPositions {
			result.AnalyzedPositions[i].Weight = result.AnalyzedPositions[i].MarketValueBase / result.TotalValue
		}
	}

	sort.Slice(result.AnalyzedPositions, func(i, j int) bool {
		return result.AnalyzedPositions[i].MarketValueBase > result.AnalyzedPositions[j].MarketValueBase
	})

	return result, nil
}

func classifyValuationIssue(err error) domain.ValuationIssueType {
	switch {
	case errors.Is(err, domain.ErrConsumedPriceNotFound):
		return domain.ValuationIssueMissingPrice
	case errors.Is(err, domain.ErrFXRateNotFound):
		return domain.ValuationIssueMissingFX
	default:
		return domain.ValuationIssueProvider
	}
}

func valuationHint(pos domain.Position) string {
	if pos.Instrument.InstrumentType == domain.InstrumentTypeFund {
		return "tick prices consume --file <file>"
	}

	return ""
}

func cleanValuationMessage(err error) string {
	if errors.Is(err, domain.ErrConsumedPriceNotFound) {
		return "Missing consumed price"
	}

	return "Missing price"
}
