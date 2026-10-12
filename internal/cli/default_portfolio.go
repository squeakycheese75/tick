package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/appdir"
)

const (
	fallbackPortfolio  = "main"
	portfolioEnvVar    = "TICK_PORTFOLIO"
	portfolioFlagUsage = "Portfolio name (default: the one set with `tick portfolio use`, else main)"
)

// addPortfolioFlag registers --portfolio on cmd and resolves it just before the command runs,
// so the stored default is read when the command runs rather than when it was built.
func addPortfolioFlag(cmd *cobra.Command, name *string) {
	cmd.Flags().StringVar(name, "portfolio", "", portfolioFlagUsage)

	previous := cmd.PreRunE
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		*name = resolvePortfolio(*name)

		if previous != nil {
			return previous(c, args)
		}

		return nil
	}
}

// resolvePortfolio picks the portfolio to act on: the flag, then TICK_PORTFOLIO, then the
// default stored by `tick portfolio use`, then "main".
func resolvePortfolio(flagValue string) string {
	for _, candidate := range []string{flagValue, os.Getenv(portfolioEnvVar), readDefaultPortfolio()} {
		if name := strings.TrimSpace(candidate); name != "" {
			return name
		}
	}

	return fallbackPortfolio
}

// readDefaultPortfolio returns the stored default, or "" when none is set or it cannot be read.
func readDefaultPortfolio() string {
	path, err := appdir.DefaultPortfolioPath()
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func writeDefaultPortfolio(name string) error {
	path, err := appdir.DefaultPortfolioPath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(name+"\n"), 0o644); err != nil {
		return fmt.Errorf("write default portfolio: %w", err)
	}

	return nil
}
