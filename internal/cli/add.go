package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/render"
)

func newAddPositionCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var qty float64
	var avgCost float64
	var quoteCurrency string
	var portfolioName string
	var instrumentType string
	var exchange string

	cmd := &cobra.Command{
		Use:   "add <symbol>",
		Short: "Add a position",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			symbol := strings.ToUpper(strings.TrimSpace(args[0]))
			if symbol == "" {
				return fmt.Errorf("symbol is required")
			}

			portfolioName = strings.TrimSpace(portfolioName)
			if portfolioName == "" {
				portfolioName = "main"
			}

			instrumentType = strings.ToLower(strings.TrimSpace(instrumentType))
			exchange = strings.ToUpper(strings.TrimSpace(exchange))
			quoteCurrency = strings.ToUpper(strings.TrimSpace(quoteCurrency))

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
				domain.AddPositionToPortfolioInput{
					PortfolioName:  portfolioName,
					Symbol:         symbol,
					InstrumentType: instrumentType,
					Exchange:       exchange,
					QuoteCurrency:  quoteCurrency,
					AvgCost:        avgCost,
					Qty:            qty,
				},
			)
			if err != nil {
				return err
			}

			return render.AddPortfolioPosition(cmd.OutOrStdout(), out.Position)
		},
	}

	cmd.Flags().Float64Var(&qty, "qty", 0, "Position quantity")
	cmd.Flags().Float64Var(&avgCost, "avg-cost", 0, "Average cost basis per unit")
	cmd.Flags().StringVar(&quoteCurrency, "quote-currency", "", "Override quote currency, e.g. USD")
	cmd.Flags().StringVar(&instrumentType, "asset-type", "", "Override asset type, e.g. equity, etf, crypto")
	cmd.Flags().StringVar(&exchange, "exchange", "", "Override exchange, e.g. NASDAQ")
	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")

	return cmd
}
