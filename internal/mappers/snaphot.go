package mappers

import (
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

func MapSnapshotPositionsToDomain(
	positions []repository.PortfolioSnapshotPosition,
) []domain.PortfolioSnapshotPosition {
	out := make([]domain.PortfolioSnapshotPosition, 0, len(positions))

	for _, p := range positions {
		out = append(out, domain.PortfolioSnapshotPosition{
			Symbol:             p.Symbol,
			Quantity:           p.Quantity,
			InstrumentCurrency: p.InstrumentCurrency,
			QuotedPrice:        p.QuotedPrice,
			FXRate:             p.FXRate,
			MarketValueBase:    p.MarketValueBase,
			Weight:             p.Weight,
		})
	}

	return out
}

func MapAnalysisToSnapshot(
	a domain.PortfolioAnalysis,
	now time.Time,
) (repository.PortfolioSnapshot, []repository.PortfolioSnapshotPosition) {
	snapshot := repository.PortfolioSnapshot{
		PortfolioName: a.PortfolioName,
		BaseCurrency:  a.BaseCurrency,
		TotalValue:    a.TotalValue,
		CapturedAt:    now,
	}

	positions := make([]repository.PortfolioSnapshotPosition, 0, len(a.AnalyzedPositions))
	for _, p := range a.AnalyzedPositions {
		positions = append(positions, repository.PortfolioSnapshotPosition{
			Symbol:             p.Symbol,
			Quantity:           p.Quantity,
			InstrumentCurrency: p.InstrumentCurrency,
			QuotedPrice:        p.QuotedPrice,
			FXRate:             p.FXRate,
			MarketValueBase:    p.MarketValueBase,
			Weight:             p.Weight,
		})
	}

	return snapshot, positions
}

func MapSnapshotToDomain(
	s repository.PortfolioSnapshot,
) domain.PortfolioSnapshot {

	return domain.PortfolioSnapshot{
		ID:            s.ID,
		PortfolioName: s.PortfolioName,
		BaseCurrency:  s.BaseCurrency,
		TotalValue:    s.TotalValue,
		CapturedAt:    s.CapturedAt,
	}
}
