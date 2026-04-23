package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/render"
)

func newNewsCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "news <ticker> [ticker...]",
		Short: "Show recent news for one or more tickers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				limit = 2
			}

			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			seen := make(map[string]struct{})
			tickers := make([]string, 0, len(args))

			for _, arg := range args {
				parts := strings.Split(arg, ",")

				for _, part := range parts {
					ticker := strings.ToUpper(strings.TrimSpace(part))
					if ticker == "" {
						continue
					}

					if _, ok := seen[ticker]; ok {
						continue
					}

					seen[ticker] = struct{}{}
					tickers = append(tickers, ticker)
				}
			}

			if len(tickers) == 0 {
				return fmt.Errorf("at least one ticker is required")
			}

			for i, ticker := range tickers {
				report, err := app.GetTickerNews.Execute(cmd.Context(), ticker, limit)
				if err != nil {
					return err
				}

				if err := render.RenderNewsItem(cmd.OutOrStdout(), report, render.NewsOptions{
					ShowLinks:      true,
					TruncateTitles: false,
				}); err != nil {
					return err
				}

				if i < len(tickers)-1 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout())
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 2, "Number of headlines per ticker")

	return cmd
}
