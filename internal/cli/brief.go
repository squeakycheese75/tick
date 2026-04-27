package cli

import (
	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/render"
)

func newBriefCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	return &cobra.Command{
		Use:   "brief",
		Short: "Show your morning market and portfolio brief",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := rt.GetMorningBrief.Execute(cmd.Context(), domain.GetMorningBriefUsecaseInput{
				PortfolioName: "main",
			})
			if err != nil {
				return err
			}

			opts := render.DefaultBriefReportOptions()

			return render.RenderBriefReport(cmd.OutOrStdout(), out.Report, opts)
		},
	}
}
