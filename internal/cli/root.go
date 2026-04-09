package cli

import (
	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/app"
)

func NewRootCmd(app *app.Runtime) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tick",
		Short: "Terminal-native portfolio and market intelligence tool",
		Long: `tick is a terminal-native portfolio and market intelligence tool.

It is designed around a simple daily workflow:
- check portfolio positions
- review news
- investigate a stock
- look at portfolio risk`,
	}

	rootCmd.AddCommand(
		newPortfolioCmd(app),
		newInfoCmd(),
		newNewsCmd(),
		newDailyCmd(app),
	)

	return rootCmd
}
