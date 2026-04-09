package cli

import (
	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/usecase"
)

func newDailyCmd(rt *app.Runtime) *cobra.Command {
	var portfolioName string
	var newsLimit int

	cmd := &cobra.Command{
		Use:   "daily",
		Short: "Show the daily portfolio brief",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := rt.GetDailyBrief.Execute(
				cmd.Context(),
				usecase.GetDailyBriefInput{
					PortfolioName: portfolioName,
					NewsLimit:     newsLimit,
				},
			)
			if err != nil {
				return err
			}

			return RenderGetDailyBrief(cmd.OutOrStdout(), out)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")
	cmd.Flags().IntVar(&newsLimit, "news-limit", 2, "Number of headlines per ticker")

	return cmd
}
