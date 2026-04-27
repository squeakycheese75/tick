package report

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/squeakycheese75/tick/internal/domain"
)

func positionValueChange(quantity, priceChange float64) float64 {
	return quantity * priceChange
}

func sortHoldingsByAbsValueChange(holding []domain.Holding) {
	sort.Slice(holding, func(i, j int) bool {
		return math.Abs(holding[i].ChangeAbsolute) > math.Abs(holding[j].ChangeAbsolute)
	})
}

func (s *ReportBuilder) getNewsSummaries(
	ctx context.Context,
	holdings domain.HoldingSummary,
	limit int,
) ([]domain.NewsSummary, error) {

	var wg sync.WaitGroup
	news := make([]domain.NewsSummary, len(holdings.Holdings))
	errCh := make(chan error, len(holdings.Holdings))

	for i, h := range holdings.Holdings {
		wg.Add(1)

		go func(i int, symbol string) {
			defer wg.Done()

			n, err := s.newsSvc.GetNews(ctx, symbol, limit)
			if err != nil {
				errCh <- err
				return
			}

			news[i] = n
		}(i, h.Symbol)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return news, nil
}

func assemblePortfolioSummary(analysis domain.PortfolioAnalysis) domain.PortfolioSummary {
	return domain.PortfolioSummary{
		Name:         analysis.PortfolioName,
		BaseCurrency: analysis.BaseCurrency,
		TotalValue:   analysis.TotalValue,
	}
}

func assembleGreeting() string {
	return "Good morning"
}

func assembleRiskSummary(risk domain.PortfolioRisk) domain.RiskSummary {
	return domain.RiskSummary{
		LargestPosition:   risk.LargestPosition,
		LargestWeight:     risk.LargestWeight,
		Top3Concentration: risk.Top3Concentration,
		Observations:      append([]string(nil), risk.Observations...),
	}
}

// func assembleAttentionSumary() domain.AttentionSummary {
// 	return domain.AttentionSummary{}
// }

func assembleHoldingSummary(positions []domain.AnalyzedPosition) (out domain.HoldingSummary) {
	var totalChange float64

	for _, pos := range positions {
		value := pos.MarketValueBase
		change := positionValueChange(pos.Quantity, pos.QuotedChange)

		out.TotalValue += value
		totalChange += change

		out.Holdings = append(out.Holdings, domain.Holding{
			Symbol:          pos.Symbol,
			Quantity:        pos.Quantity,
			MarketValueBase: value,
			Weight:          pos.Weight,
			QuotedPrice:     pos.QuotedPrice,
			ChangePercent:   pos.QuotedChangePct,
			ChangeAbsolute:  change,
			PriceCurrency:   pos.PriceCurrency,
		})
	}

	out.Change = &domain.ValueChangeSummary{
		Absolute: totalChange,
	}

	previousValue := out.TotalValue - totalChange
	if previousValue > 0 {
		out.Change.Percent = totalChange / previousValue
	}

	return out
}
