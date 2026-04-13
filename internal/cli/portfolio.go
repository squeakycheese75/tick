package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/usecase"
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
		newPortfolioAddPositionCmd(runtimeBuilder),
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

func newPortfolioAddPositionCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var qty float64
	var avgCost float64
	var quoteCurrency string
	var portfolioName string
	var assetType string
	var exchange string

	cmd := &cobra.Command{
		Use:   "add <symbol>",
		Short: "Add a position to a portfolio",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := strings.ToUpper(strings.TrimSpace(args[0]))
			if symbol == "" {
				return fmt.Errorf("symbol is required")
			}

			portfolioName = strings.TrimSpace(portfolioName)
			if portfolioName == "" {
				return fmt.Errorf("portfolio is required")
			}

			assetType = strings.ToLower(strings.TrimSpace(assetType))
			if assetType == "" {
				return fmt.Errorf("asset-type is required")
			}

			exchange = strings.ToUpper(strings.TrimSpace(exchange))
			if exchange == "" {
				return fmt.Errorf("exchange is required")
			}

			quoteCurrency = strings.ToUpper(strings.TrimSpace(quoteCurrency))
			if quoteCurrency == "" {
				return fmt.Errorf("quote-currency is required")
			}

			if qty <= 0 {
				return fmt.Errorf("qty must be greater than 0")
			}

			if avgCost < 0 {
				return fmt.Errorf("avg-cost must be 0 or greater")
			}

			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := app.AddPosition.Execute(
				cmd.Context(),
				usecase.AddPositionToPortfolioInput{
					PortfolioName: portfolioName,
					Symbol:        symbol,
					AssetType:     assetType,
					Exchange:      exchange,
					QuoteCurrency: quoteCurrency,
					AvgCost:       avgCost,
					Qty:           qty,
				},
			)
			if err != nil {
				return err
			}

			return RenderAddPortfolioPosition(cmd.OutOrStdout(), *out)
		},
	}

	cmd.Flags().Float64Var(&qty, "qty", 0, "Position quantity")
	cmd.Flags().Float64Var(&avgCost, "avg-cost", 0, "Average cost basis per unit")
	cmd.Flags().StringVar(&quoteCurrency, "quote-currency", "", "Instrument quote currency, e.g. USD")
	cmd.Flags().StringVar(&assetType, "asset-type", "", "Instrument asset type, e.g. equity, etf, crypto")
	cmd.Flags().StringVar(&exchange, "exchange", "", "Instrument exchange, e.g. NASDAQ")
	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")

	return cmd
}

func newPortfolioCreateCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
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

			app, err := runtimeBuilder()
			if err != nil {
				return err
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

			var in usecase.ImportPortfolioInput
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
