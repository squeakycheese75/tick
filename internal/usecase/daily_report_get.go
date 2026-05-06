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
