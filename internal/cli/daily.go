package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/render"
)

func newDailyCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var portfolioName string
	var newsLimit int
	var ai bool

	cmd := &cobra.Command{
		Use:   "daily",
		Short: "Show the daily portfolio brief",
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
				},
			)
			if err != nil {
				return err
			}

			truncate, _ := cmd.Flags().GetBool("truncate")
			newsCount, _ := cmd.Flags().GetInt("news-count")
			showLinks, _ := cmd.Flags().GetBool("links")

			opts := render.DefaultDailyReportOptions()
			opts.News.TruncateTitles = truncate
			opts.News.MaxHeadlines = newsCount
			opts.News.ShowLinks = showLinks

			return render.DailyReport(cmd.OutOrStdout(), out, opts)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")
	cmd.Flags().BoolVar(&ai, "ai", false, "Include AI analysis")
	cmd.Flags().Bool("truncate", false, "Hide full details")
	cmd.Flags().Int("news-count", 1, "Number of headlines per ticker")
	cmd.Flags().Bool("links", false, "Show headline URLs")

	return cmd
}
