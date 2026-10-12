package render

import (
	"os"
	"time"

	"github.com/mattn/go-isatty"
)

type DailyReportOptions struct {
	// Short shows movers and news only.
	Short         bool
	ShowAttention bool
	Summary       SummaryOptions
	Holdings      HoldingsOptions
	Risk          RiskOptions
	News          NewsOptions
	AI            AIOptions
}
type SummaryOptions struct {
	ShowSnapshotDelta bool
	HideZeroDelta     bool
	// Now is the time prices are judged against when deciding whether to say they are old.
	Now time.Time
}

type HoldingsOptions struct {
	Title             string
	ShowSnapshotDelta bool
	HideZeroDelta     bool
	Color             bool
	ShowTop           int
}

type RiskOptions struct {
	Compact          bool
	ShowObservations bool
}

type NewsOptions struct {
	MaxHeadlines   int
	ShowLinks      bool
	TruncateTitles bool
	HeadlineMaxLen int
}

type AIOptions struct {
	Show bool
}

func DefaultDailyReportOptions() DailyReportOptions {
	return DailyReportOptions{

		Summary: SummaryOptions{
			ShowSnapshotDelta: true,
			HideZeroDelta:     true,
			Now:               time.Now(),
		},
		Holdings: HoldingsOptions{
			Title:             "Holdings",
			ShowSnapshotDelta: true,
			HideZeroDelta:     true,
			Color:             isatty.IsTerminal(os.Stdout.Fd()),
		},
		Risk: RiskOptions{
			Compact:          true,
			ShowObservations: false,
		},
		News: NewsOptions{
			MaxHeadlines:   1,
			ShowLinks:      false,
			TruncateTitles: true,
			HeadlineMaxLen: 100,
		},
		AI: AIOptions{
			Show: false,
		},
		ShowAttention: false,
	}
}

type PortfolioSummaryOptions struct {
	ShowHeader    bool
	ShowTotals    bool
	ShowPositions bool
	ShowPrices    bool
	Color         bool
	Now           time.Time
}

func DefaultPortfolioSummaryOptions() PortfolioSummaryOptions {
	return PortfolioSummaryOptions{
		ShowHeader:    true,
		ShowTotals:    true,
		ShowPositions: true,
		Color:         isatty.IsTerminal(os.Stdout.Fd()),
		Now:           time.Now(),
	}
}

type PortfolioRiskOptions struct {
	ShowObservations bool
}

func DefaultPortfolioRiskOptions() PortfolioRiskOptions {
	return PortfolioRiskOptions{
		ShowObservations: true,
	}
}

// ShortDailyReportOptions is the quick look: movers and news, without the snapshot comparison.
func ShortDailyReportOptions() DailyReportOptions {
	opts := DefaultDailyReportOptions()
	opts.Short = true
	opts.Summary.ShowSnapshotDelta = false
	opts.Holdings.Title = "Movers"

	return opts
}
