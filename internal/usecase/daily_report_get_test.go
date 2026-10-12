package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/report"
	"github.com/squeakycheese75/tick/internal/usecase"
)

type fakeReportBuilder struct {
	got report.BuildDailyReportParams
	err error
}

func (f *fakeReportBuilder) BuildDailyReport(_ context.Context, in report.BuildDailyReportParams) (domain.DailyReport, error) {
	f.got = in
	return domain.DailyReport{Portfolio: domain.PortfolioSummary{Name: in.PortfolioName}}, f.err
}

type fakeSummarizer struct {
	enabled bool
	err     error
}

func (f fakeSummarizer) Enabled() bool { return f.enabled }
func (f fakeSummarizer) Summarize(context.Context, domain.DailyReport) (string, error) {
	return "summary", f.err
}

func TestGetDailyReport_Execute_ShortDoesNotSaveASnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		short        bool
		wantSnapshot bool
	}{
		{"full report saves a snapshot", false, true},
		{"short report leaves history alone", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := &fakeReportBuilder{}
			uc := usecase.NewGetDailyReportUseCase(builder, fakeSummarizer{}, nil)

			if _, err := uc.Execute(context.Background(), domain.GetDailyReportInput{PortfolioName: "trending", Short: tt.short}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if builder.got.SaveSnapshot != tt.wantSnapshot || builder.got.Short != tt.short || builder.got.PortfolioName != "trending" {
				t.Fatalf("builder params = %+v", builder.got)
			}
		})
	}
}

func TestGetDailyReport_Execute_AISummary(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	tests := []struct {
		name       string
		builderErr error
		summarizer fakeSummarizer
		withAI     bool
		wantAI     string
		wantErr    string
		// wantUnavailable: a model that cannot be reached is a provider failure, not a bug.
		wantUnavailable bool
	}{
		{name: "no AI asked for", summarizer: fakeSummarizer{enabled: true}},
		{name: "AI summary attached", summarizer: fakeSummarizer{enabled: true}, withAI: true, wantAI: "summary"},
		{name: "AI asked for but not configured", withAI: true, wantErr: "ai is not configured"},
		{name: "AI failure is returned", summarizer: fakeSummarizer{enabled: true, err: boom}, withAI: true, wantErr: "boom", wantUnavailable: true},
		{name: "report failure is returned", builderErr: boom, wantErr: "boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := usecase.NewGetDailyReportUseCase(&fakeReportBuilder{err: tt.builderErr}, tt.summarizer, nil)

			out, err := uc.Execute(context.Background(), domain.GetDailyReportInput{WithAI: tt.withAI})

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr && !errors.Is(err, boom) {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				if tt.wantUnavailable && domain.KindOf(err) != domain.KindUnavailable {
					t.Fatalf("kind = %s, want unavailable: a model that cannot be reached is not a bug", domain.KindOf(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.AISummary != tt.wantAI {
				t.Fatalf("AI summary = %q, want %q", out.AISummary, tt.wantAI)
			}
		})
	}
}
