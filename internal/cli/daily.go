package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/render"
)

func newDailyCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	return buildDailyCmd(runtimeBuilder, "daily", "Show the daily portfolio report", false)
}

// newBriefCmd keeps `tick brief` working as a hidden alias for `tick daily --short`.
func newBriefCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	cmd := buildDailyCmd(runtimeBuilder, "brief", "Alias for `daily --short`", true)
	cmd.Hidden = true

	return cmd
}

func buildDailyCmd(runtimeBuilder RuntimeBuilder, use, summary string, alwaysShort bool) *cobra.Command {
	var portfolioName string
	var newsLimit int
	var ai bool
	short := alwaysShort

	cmd := &cobra.Command{
		Use:   use,
		Short: summary,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeBuilder()
			if err != nil {
				return err
			}

			if ai && rt.GetDailyReport == nil {
				return fmt.Errorf("ai is not configured")
			}

			out, err := rt.GetDailyReport.Execute(
				cmd.Context(),
				domain.GetDailyReportInput{
					PortfolioName: portfolioName,
					NewsLimit:     newsLimit,
					WithAI:        ai,
					Short:         short,
				},
			)
			if err != nil {
				return err
			}

			truncate, _ := cmd.Flags().GetBool("truncate")
			newsCount, _ := cmd.Flags().GetInt("news-count")
			showLinks, _ := cmd.Flags().GetBool("links")

			opts := render.DefaultDailyReportOptions()
			if short {
				opts = render.ShortDailyReportOptions()
			}
			opts.News.TruncateTitles = truncate
			opts.News.MaxHeadlines = newsCount
			opts.News.ShowLinks = showLinks

			return render.DailyReport(cmd.OutOrStdout(), out, opts)
		},
	}

	addPortfolioFlag(cmd, &portfolioName)
	cmd.Flags().BoolVar(&ai, "ai", false, "Include AI analysis")
	cmd.Flags().Bool("truncate", false, "Hide full details")
	cmd.Flags().Int("news-count", 1, "Number of headlines per ticker")
	cmd.Flags().Bool("links", false, "Show headline URLs")

	if !alwaysShort {
		cmd.Flags().BoolVar(&short, "short", false, "Movers and news only; does not save a snapshot")
	}

	return cmd
}
