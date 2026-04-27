package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/report"
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
	in domain.GetDailyReportInput,
) (domain.GetDailyReportOutput, error) {
	in.ApplyDefaults()

	dailyReport, err := uc.reportBuilder.BuildDailyReport(ctx, report.BuildDailyReportParams{
		PortfolioName: in.PortfolioName,
		NewsLimit:     in.NewsLimit,
		SaveSnapshot:  true,
	})
	if err != nil {
		return domain.GetDailyReportOutput{}, err
	}

	out := domain.GetDailyReportOutput{
		DailyReport: dailyReport,
	}

	if !in.WithAI {
		return out, nil
	}

	if !uc.summarizer.Enabled() {
		return domain.GetDailyReportOutput{}, fmt.Errorf("ai not configured")
	}

	summary, err := uc.summarizer.Summarize(ctx, dailyReport)
	if err != nil {
		return domain.GetDailyReportOutput{}, err
	}

	out.AISummary = summary
	return out, nil
}

// func (uc *GetDailyReportUseCase) Execute(
// 	ctx context.Context,
// 	in domain.GetDailyReportInput,
// ) (domain.GetDailyReportOutput, error) {
// 	in.ApplyDefaults()

// 	dailyReport, err := uc.reportBuilder.BuildDailyReport(
// 		ctx,
// 		report.BuildDailyReportParams{
// 			PortfolioName: in.PortfolioName,
// 			NewsLimit:     in.NewsLimit,
// 		},
// 	)
// 	if err != nil {
// 		return domain.GetDailyReportOutput{}, err
// 	}

// 	snapshot, positions := snapshot.MapAnalysisToSnapshot(dailyReport.Analysis, time.Now())

// 	_, err = uc.snapshotRepo.Create(ctx, snapshot, positions)
// 	if err != nil {
// 		return domain.GetDailyReportOutput{}, fmt.Errorf("save portfolio snapshot: %w", err)
// 	}

// 	out := domain.GetDailyReportOutput{
// 		DailyReport: dailyReport.Report,
// 	}

// 	out.DailyReport, err = uc.enrichWithPreviousSnapshot(
// 		ctx,
// 		out.DailyReport,
// 		snapshot.PortfolioName,
// 		snapshot.CapturedAt,
// 	)
// 	if err != nil {
// 		return domain.GetDailyReportOutput{}, err
// 	}

// 	if in.WithAI && !uc.summarizer.Enabled() {
// 		return domain.GetDailyReportOutput{}, fmt.Errorf("ai not configured")
// 	}

// 	if !in.WithAI {
// 		return out, nil
// 	}

// 	summary, err := uc.summarizer.Summarize(ctx, dailyReport.Report)
// 	if err != nil {
// 		return domain.GetDailyReportOutput{}, err
// 	}

// 	out.AISummary = summary

// 	return out, nil
// }

// func (uc *GetDailyReportUseCase) enrichWithPreviousSnapshot(
// 	ctx context.Context,
// 	dailyReport domain.DailyReport,
// 	portfolioName string,
// 	capturedAt time.Time,
// ) (domain.DailyReport, error) {
// 	prev, err := uc.snapshotRepo.GetLatestBefore(ctx, portfolioName, capturedAt)
// 	if err != nil {
// 		if errors.Is(err, domain.ErrPortfolioSnapshotNotFound) {
// 			return dailyReport, nil
// 		}

// 		return domain.DailyReport{}, fmt.Errorf("get previous snapshot: %w", err)
// 	}

// 	prevPositions, err := uc.snapshotRepo.ListPositionsBySnapshotID(ctx, prev.ID)
// 	if err != nil {
// 		return domain.DailyReport{}, fmt.Errorf("list previous snapshot positions: %w", err)
// 	}

// 	return report.EnrichDailyReportWithSnapshot(
// 		dailyReport,
// 		snapshot.MapSnapshotToDomain(prev),
// 		snapshot.MapSnapshotPositionsToDomain(prevPositions),
// 	), nil
// }
