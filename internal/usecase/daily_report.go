package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/squeakycheese75/tick/internal/domain/analysis"
	"github.com/squeakycheese75/tick/internal/repository"
)

type GetDailyReportUseCase struct {
	reportBuilder ReportBuilder
	summarizer    DailyReportSummarizer
	snapshotRepo  PortfolioSnapshotRepository
}

func NewGetDailyReportUseCase(reportSvc ReportBuilder, summarizer DailyReportSummarizer, snapshotRepo PortfolioSnapshotRepository) *GetDailyReportUseCase {
	return &GetDailyReportUseCase{
		reportBuilder: reportSvc,
		summarizer:    summarizer,
		snapshotRepo:  snapshotRepo,
	}
}

func (uc *GetDailyReportUseCase) Execute(
	ctx context.Context,
	in GetDailyReportInput,
) (GetDailyReportOutput, error) {

	if in.PortfolioName == "" {
		in.PortfolioName = "main"
	}

	if in.NewsLimit <= 0 {
		in.NewsLimit = 2
	}

	dailyReport, err := uc.reportBuilder.BuildDailyReport(
		ctx,
		in.PortfolioName,
		in.NewsLimit,
	)
	if err != nil {
		return GetDailyReportOutput{}, err
	}

	snapshot, positions := mapAnalysisToSnapshot(dailyReport.Analysis, time.Now())

	_, err = uc.snapshotRepo.Create(ctx, snapshot, positions)
	if err != nil {
		return GetDailyReportOutput{}, fmt.Errorf("save portfolio snapshot: %w", err)
	}

	out := GetDailyReportOutput{
		DailyReport: dailyReport.Report,
	}

	if in.WithAI && !uc.summarizer.Enabled() {
		return GetDailyReportOutput{}, fmt.Errorf("ai not configured")
	}

	if !in.WithAI {
		return out, nil
	}

	summary, err := uc.summarizer.Summarize(ctx, dailyReport.Report)
	if err != nil {
		return GetDailyReportOutput{}, err
	}

	out.AISummary = summary

	return out, nil
}

func mapAnalysisToSnapshot(
	a analysis.PortfolioAnalysis,
	now time.Time,
) (repository.CreatePortfolioSnapshot, []repository.CreatePortfolioSnapshotPosition) {
	snapshot := repository.CreatePortfolioSnapshot{
		PortfolioName: a.PortfolioName,
		BaseCurrency:  a.BaseCurrency,
		TotalValue:    a.TotalValue,
		CapturedAt:    now,
	}

	positions := make([]repository.CreatePortfolioSnapshotPosition, 0, len(a.AnalyzedPositions))
	for _, p := range a.AnalyzedPositions {
		positions = append(positions, repository.CreatePortfolioSnapshotPosition{
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
