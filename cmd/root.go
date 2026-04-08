package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
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
		newPortfolioCmd(),
		newInfoCmd(),
		newNewsCmd(),
	)

	return rootCmd
}
