package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/usecase"
)

func newPortfolioCmd(app *app.Runtime) *cobra.Command {
	portfolioCmd := &cobra.Command{
		Use:   "portfolio",
		Short: "Portfolio commands",
	}

	portfolioCmd.AddCommand(
		newPortfolioCreateCmd(app),
		newPortfolioSummaryCmd(app),
		newPortfolioRiskCmd(app),
		newPortfolioAddPositionCmd(app),
	)

	return portfolioCmd
}

func newPortfolioSummaryCmd(app *app.Runtime) *cobra.Command {
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show portfolio summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := app.GetPortfolioSummary.Execute(
				cmd.Context(),
				usecase.GetPortfolioSummaryUsecaseInput{
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

func newPortfolioAddPositionCmd(app *app.Runtime) *cobra.Command {
	var qty float64
	var avgCost float64
	var currency string
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "add <ticker>",
		Short: "Add or update a portfolio position",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ticker := strings.ToUpper(strings.TrimSpace(args[0]))
			if ticker == "" {
				return fmt.Errorf("ticker is required")
			}

			if qty <= 0 {
				return fmt.Errorf("qty must be greater than 0")
			}

			if avgCost < 0 {
				return fmt.Errorf("avg-cost must be 0 or greater")
			}

			currency = strings.ToUpper(strings.TrimSpace(currency))
			if currency == "" {
				currency = "USD"
			}

			out, err := app.AddPosition.Execute(
				cmd.Context(),
				usecase.AddPositionToPortfolioUseCaseInput{
					PortfolioName: portfolioName,
					Ticker:        ticker,
					Currency:      currency,
					AvgCost:       avgCost,
					Qty:           qty,
				},
			)
			if err != nil {
				return err
			}

			return RenderAddPortfolioPosition(cmd.OutOrStdout(), out)
		},
	}

	cmd.Flags().Float64Var(&qty, "qty", 0, "Position quantity")
	cmd.Flags().Float64Var(&avgCost, "avg-cost", 0, "Average cost basis per unit")
	cmd.Flags().StringVar(&currency, "currency", "USD", "Position currency")
	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")

	return cmd
}

func newPortfolioCreateCmd(app *app.Runtime) *cobra.Command {
	var baseCurrency string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create or update a portfolio",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("portfolio name is required")
			}

			baseCurrency = strings.ToUpper(strings.TrimSpace(baseCurrency))
			if baseCurrency == "" {
				return fmt.Errorf("base-currency is required")
			}

			out, err := app.CreatePortfolio.Execute(
				cmd.Context(),
				usecase.CreatePortfolioUsecaseInput{
					PortfolioName: name,
					BaseCurrency:  baseCurrency,
				},
			)
			if err != nil {
				return err
			}

			return RenderCreatePortfolio(cmd.OutOrStdout(), *out)
		},
	}

	cmd.Flags().StringVar(
		&baseCurrency,
		"base-currency",
		"EUR",
		"Base currency for the portfolio (e.g. EUR, USD, GBP)",
	)

	return cmd
}

func newPortfolioRiskCmd(rt *app.Runtime) *cobra.Command {
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "risk",
		Short: "Show portfolio concentration and basic risk summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := rt.GetPortfolioRisk.Execute(
				cmd.Context(),
				usecase.GetPortfolioRiskInput{
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
