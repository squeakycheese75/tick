package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
	snapshot "github.com/squeakycheese75/tick/internal/mappers"
	"github.com/squeakycheese75/tick/internal/repository"
)

type SnapRepository interface {
	Create(ctx context.Context, in repository.PortfolioSnapshot, positions []repository.PortfolioSnapshotPosition) (int64, error)
	GetLatestBefore(ctx context.Context, portfolioName string, before time.Time) (repository.PortfolioSnapshot, error)
	ListPositionsBySnapshotID(ctx context.Context, snapshotID int64) ([]repository.PortfolioSnapshotPosition, error)
}

type SnapshotService struct {
	snapshots SnapRepository
}

func NewSnapshotService(repo SnapRepository) *SnapshotService {
	return &SnapshotService{
		snapshots: repo,
	}
}

func (s *SnapshotService) SaveAndEnrichDailyReport(
	ctx context.Context,
	dailyReport domain.DailyReport,
	analysis domain.PortfolioAnalysis,
) (domain.DailyReport, error) {
	currentSnapshot, currentPositions := snapshot.MapAnalysisToSnapshot(analysis, time.Now())

	_, err := s.snapshots.Create(ctx, currentSnapshot, currentPositions)
	if err != nil {
		return domain.DailyReport{}, fmt.Errorf("save portfolio snapshot: %w", err)
	}

	previousSnapshot, err := s.snapshots.GetLatestBefore(
		ctx,
		currentSnapshot.PortfolioName,
		currentSnapshot.CapturedAt,
	)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioSnapshotNotFound) {
			return dailyReport, nil
		}

		return domain.DailyReport{}, fmt.Errorf("get previous snapshot: %w", err)
	}

	previousPositions, err := s.snapshots.ListPositionsBySnapshotID(ctx, previousSnapshot.ID)
	if err != nil {
		return domain.DailyReport{}, fmt.Errorf("list previous snapshot positions: %w", err)
	}

	return EnrichDailyReportWithSnapshot(
		dailyReport,
		snapshot.MapSnapshotToDomain(previousSnapshot),
		snapshot.MapSnapshotPositionsToDomain(previousPositions),
	), nil
}

func EnrichDailyReportWithSnapshot(
	dailyReport domain.DailyReport,
	previousSnapshot domain.PortfolioSnapshot,
	previousPositions []domain.PortfolioSnapshotPosition,
) domain.DailyReport {
	delta := dailyReport.Portfolio.TotalValue - previousSnapshot.TotalValue

	var pct float64
	if previousSnapshot.TotalValue != 0 {
		pct = delta / previousSnapshot.TotalValue
	}

	dailyReport.Portfolio.Change = &domain.ValueChangeSummary{
		Absolute: delta,
		Percent:  pct,
	}

	previousBySymbol := make(map[string]domain.PortfolioSnapshotPosition, len(previousPositions))
	for _, p := range previousPositions {
		previousBySymbol[p.Symbol] = p
	}

	for i := range dailyReport.TopHoldings.Holdings {
		current := dailyReport.TopHoldings.Holdings[i]
		previous, ok := previousBySymbol[current.Symbol]
		if !ok {
			continue
		}

		valueDelta := current.MarketValueBase - previous.MarketValueBase

		var valuePct float64
		if previous.MarketValueBase != 0 {
			valuePct = valueDelta / previous.MarketValueBase
		}

		dailyReport.TopHoldings.Holdings[i].SinceLastSnapshot = &domain.ValueChangeSummary{
			Absolute: valueDelta,
			Percent:  valuePct,
		}
	}

	return dailyReport
}
