package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/domain"
)

func newPortfolioCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	portfolioCmd := &cobra.Command{
		Use:   "portfolio",
		Short: "Portfolio commands",
	}

	portfolioCmd.AddCommand(
		newPortfolioCreateCmd(runtimeBuilder),
		newPortfolioSummaryCmd(runtimeBuilder),
		newPortfolioRiskCmd(runtimeBuilder),
		newPortfolioImportCmd(runtimeBuilder),
	)

	return portfolioCmd
}

func newPortfolioSummaryCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show portfolio summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := app.GetPortfolioSummary.Execute(
				cmd.Context(),
				domain.GetPortfolioSummaryUsecaseInput{
					PortfolioName: portfolioName,
				},
			)
			if err != nil {
				return err
			}

			return RenderGetPortfolioSummary(cmd.OutOrStdout(), out)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")
	return cmd
}

func newPortfolioCreateCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var base string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create or update a portfolio",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("portfolio name is required")
			}

			base = strings.ToUpper(strings.TrimSpace(base))
			if base == "" {
				return fmt.Errorf("base currency is required")
			}

			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := app.CreatePortfolio.Execute(
				cmd.Context(),
				domain.CreatePortfolioUsecaseInput{
					PortfolioName: name,
					BaseCurrency:  base,
				},
			)
			if err != nil {
				return err
			}

			return RenderCreatePortfolio(cmd.OutOrStdout(), *out)
		},
	}

	cmd.Flags().StringVar(
		&base,
		"base",
		"EUR",
		"Base currency for the portfolio (e.g. EUR, USD, GBP)",
	)

	return cmd
}

func newPortfolioRiskCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "risk",
		Short: "Show portfolio concentration and basic risk summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := app.GetPortfolioRisk.Execute(
				cmd.Context(),
				domain.GetPortfolioRiskInput{
					PortfolioName: portfolioName,
				},
			)
			if err != nil {
				return err
			}

			return RenderGetPortfolioRisk(cmd.OutOrStdout(), out)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")
	return cmd
}

func newPortfolioImportCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a portfolio from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath = strings.TrimSpace(filePath)
			if filePath == "" {
				return fmt.Errorf("file is required")
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file %q: %w", filePath, err)
			}

			var in domain.ImportPortfolioInput
			if err := json.Unmarshal(data, &in); err != nil {
				return fmt.Errorf("decode import file %q: %w", filePath, err)
			}

			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := app.ImportPortfolio.Execute(cmd.Context(), in)
			if err != nil {
				return err
			}

			return renderImportPortfolio(cmd.OutOrStdout(), *out)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to portfolio JSON file")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
