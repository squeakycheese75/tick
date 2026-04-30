package report

import "github.com/squeakycheese75/tick/internal/domain"

type TargetEvaluator struct{}

func NewTargetEvaluator() *TargetEvaluator {
	return &TargetEvaluator{}
}

func (e *TargetEvaluator) Evaluate(target domain.Target, quote domain.Quote) domain.TargetStatus {
	distancePct := ((target.TargetPrice - quote.Price) / quote.Price) * 100

	hit := false

	if target.Type == domain.TargetTypeTakeProfit {
		hit = quote.Price >= target.TargetPrice
	}

	if target.Type == domain.TargetTypeStopLoss {
		hit = quote.Price <= target.TargetPrice
	}

	return domain.TargetStatus{
		Symbol:       target.Symbol,
		Type:         target.Type,
		CurrentPrice: quote.Price,
		TargetPrice:  target.TargetPrice,
		Currency:     target.QuoteCurrency,
		Hit:          hit,
		DistancePct:  distancePct,
	}
}
