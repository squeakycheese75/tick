package usecase

import (
	"context"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/report"
)

type GetMorningBriefUsecase struct {
	reportBuilder ReportBuilder
}

func NewGetMorningBriefUsecase(reportBuilder ReportBuilder) *GetMorningBriefUsecase {
	return &GetMorningBriefUsecase{
		reportBuilder: reportBuilder,
	}
}

func (uc *GetMorningBriefUsecase) Execute(ctx context.Context, in domain.GetMorningBriefUsecaseInput) (domain.GetMorningBriefUsecaseOutput, error) {
	in.ApplyDefaults()

	report, err := uc.reportBuilder.BuildMorningBriefReport(ctx, report.BuildMorningBriefReportParams{
		PortfolioName: in.PortfolioName,
	})
	if err != nil {
		return domain.GetMorningBriefUsecaseOutput{}, err
	}

	return domain.GetMorningBriefUsecaseOutput{
		Report: report,
	}, nil
}
