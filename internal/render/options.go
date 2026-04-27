package render

import (
	"os"

	"github.com/mattn/go-isatty"
)

type DailyReportOptions struct {
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
}

type HoldingsOptions struct {
	Title             string
	ShowSnapshotDelta bool
	HideZeroDelta     bool
	Color             bool
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
		},
		Holdings: HoldingsOptions{
			Title:             "Holdings",
			ShowSnapshotDelta: true,
			HideZeroDelta:     true,
			Color:             true,
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
	Color         bool
}

func DefaultPortfolioSummaryOptions() PortfolioSummaryOptions {
	return PortfolioSummaryOptions{
		ShowHeader:    true,
		ShowTotals:    true,
		ShowPositions: true,
		Color:         isatty.IsTerminal(os.Stdout.Fd()),
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

type BriefReportOptions struct {
	Summary  SummaryOptions
	Holdings HoldingsOptions
	News     NewsOptions
}

func DefaultBriefReportOptions() BriefReportOptions {
	color := isatty.IsTerminal(os.Stdout.Fd())

	return BriefReportOptions{
		Summary: SummaryOptions{
			ShowSnapshotDelta: false,
			HideZeroDelta:     true,
		},
		Holdings: HoldingsOptions{
			Title:             "Movers",
			ShowSnapshotDelta: true,
			HideZeroDelta:     true,
			Color:             color,
		},
		News: NewsOptions{
			MaxHeadlines:   1,
			ShowLinks:      false,
			TruncateTitles: true,
			HeadlineMaxLen: 100,
		},
	}
}
