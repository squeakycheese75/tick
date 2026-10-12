package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/render"
)

func newTradesCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	tradesCmd := &cobra.Command{
		Use:   "trades",
		Short: "Trade commands",
	}

	tradesCmd.AddCommand(newTradesImportCmd(runtimeBuilder))

	return tradesCmd
}

func newTradesImportCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var filePath string
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Apply trades from a CSV file to a portfolio",
		Long: "Apply buys and sells from a CSV file to a portfolio's positions.\n\n" +
			"Required columns: date, symbol, side (buy|sell), quantity, price, currency.\n" +
			"Optional columns: fee, fx_rate, ref, exchange, instrument_type, quote_currency.\n\n" +
			"Importing the same file again is safe: trades already recorded are skipped.",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath = strings.TrimSpace(filePath)
			if filePath == "" {
				return fmt.Errorf("file is required")
			}

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open file %q: %w", filePath, err)
			}
			defer func() { _ = file.Close() }()

			trades, err := parseTradesCSV(file)
			if err != nil {
				return fmt.Errorf("parse trades file %q: %w", filePath, err)
			}

			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := app.ImportTrades.Execute(cmd.Context(), domain.ImportTradesInput{
				PortfolioName: strings.TrimSpace(portfolioName),
				Trades:        trades,
			})
			if err != nil {
				return err
			}

			return render.ImportTrades(cmd.OutOrStdout(), *out)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to trades CSV file")
	addPortfolioFlag(cmd, &portfolioName)
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
