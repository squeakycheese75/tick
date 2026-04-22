package cli

import (
	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/app"
)

type RuntimeBuilder func() (*app.Runtime, error)

func NewRootCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "tick",
	}

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newDailyCmd(runtimeBuilder))
	rootCmd.AddCommand(newPortfolioCmd(runtimeBuilder))
	rootCmd.AddCommand(newAddPositionCmd(runtimeBuilder))
	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newNewsCmd(runtimeBuilder))
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newDashboardCmd(runtimeBuilder))

	return rootCmd
}
