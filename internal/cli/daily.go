package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/usecase"
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
				usecase.GetDailyReportInput{
					PortfolioName: portfolioName,
					NewsLimit:     newsLimit,
					WithAI:        ai,
				},
			)
			if err != nil {
				return err
			}

			return RenderGetDailyReport(cmd.OutOrStdout(), out)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")
	cmd.Flags().IntVar(&newsLimit, "news-limit", 2, "Number of headlines per ticker")
	cmd.Flags().BoolVar(&ai, "ai", false, "Include AI analysis")

	return cmd
}
