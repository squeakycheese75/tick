package report_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/report"
)

type fakeAnalysis struct {
	analysis domain.PortfolioAnalysis
	err      error
}

func (f fakeAnalysis) GetAnalysis(context.Context, string) (domain.PortfolioAnalysis, error) {
	return f.analysis, f.err
}

// fakeNews is called from one goroutine per holding, so it guards what it records.
type fakeNews struct {
	err   error
	mu    sync.Mutex
	asked []string
}

func (f *fakeNews) GetNews(_ context.Context, ticker string, _ int) (domain.NewsSummary, error) {
	f.mu.Lock()
	f.asked = append(f.asked, ticker)
	f.mu.Unlock()

	return domain.NewsSummary{Ticker: ticker}, f.err
}

// A short report must not touch risk, targets, insights or snapshots, so those are left nil:
// reaching any of them would panic.
func newShortBuilder(a fakeAnalysis, news *fakeNews) *report.ReportBuilder {
	return report.NewReportBuilder(a, nil, nil, news, nil, nil, nil)
}

func TestBuildDailyReport_Short_RanksMoversByTheSizeOfTheirMove(t *testing.T) {
	t.Parallel()

	a := fakeAnalysis{analysis: domain.PortfolioAnalysis{
		PortfolioName: "p", BaseCurrency: "EUR", TotalValue: 300,
		AnalyzedPositions: []domain.AnalyzedPosition{
			{Symbol: "SMALL", Quantity: 1, QuotedChange: 1, MarketValueBase: 100},
			{Symbol: "DOWN", Quantity: 1, QuotedChange: -9, MarketValueBase: 100},
			{Symbol: "UP", Quantity: 1, QuotedChange: 5, MarketValueBase: 100},
		},
		ValuationIssues: []domain.ValuationIssue{{Symbol: "NOPRICE"}},
	}}
	news := &fakeNews{}

	got, err := newShortBuilder(a, news).BuildDailyReport(context.Background(), report.BuildDailyReportParams{
		PortfolioName: "p", NewsLimit: 1, Short: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var order []string
	for _, h := range got.TopHoldings.Holdings {
		order = append(order, h.Symbol)
	}
	if len(order) != 3 || order[0] != "DOWN" || order[1] != "UP" || order[2] != "SMALL" {
		t.Fatalf("movers = %v, want DOWN, UP, SMALL (largest move first, up or down)", order)
	}
	if len(got.News) != 3 || len(news.asked) != 3 {
		t.Fatalf("expected news for each holding, got %d (asked %v)", len(got.News), news.asked)
	}
	if len(got.ValuationIssues) != 1 || got.Portfolio.Name != "p" {
		t.Fatalf("report = %+v", got)
	}
}

func TestBuildDailyReport_Short_Errors(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	withPosition := domain.PortfolioAnalysis{AnalyzedPositions: []domain.AnalyzedPosition{{Symbol: "A"}}}

	tests := []struct {
		name     string
		analysis fakeAnalysis
		news     *fakeNews
	}{
		{"analysis fails", fakeAnalysis{err: boom}, &fakeNews{}},
		{"news fails", fakeAnalysis{analysis: withPosition}, &fakeNews{err: boom}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := newShortBuilder(tt.analysis, tt.news).BuildDailyReport(context.Background(), report.BuildDailyReportParams{Short: true})
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
		})
	}
}
