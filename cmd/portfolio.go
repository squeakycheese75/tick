package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/adapters/market"
	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/cli"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/store"
)

func newPortfolioCmd() *cobra.Command {
	portfolioCmd := &cobra.Command{
		Use:   "portfolio",
		Short: "Portfolio commands",
	}

	portfolioCmd.AddCommand(
		newPortfolioCreateCmd(),
		newPortfolioSummaryCmd(),
		newPortfolioRiskCmd(),
		newPortfolioAddPositionCmd(),
	)

	return portfolioCmd
}

func newPortfolioSummaryCmd() *cobra.Command {
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show current portfolio positions and allocation summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := store.Open("tick.db")
			if err != nil {
				return err
			}
			defer db.Close()

			portfolioRepo := store.NewPortfolioRepository(db)
			positionRepo := store.NewPositionRepository(db)
			priceProvider := market.NewStaticPriceProvider()
			fxProvider := market.NewStaticFXProvider()

			service := app.NewPortfolioService(portfolioRepo, positionRepo, priceProvider, fxProvider)

			summary, err := service.GetSummary(cmd.Context(), portfolioName)
			if err != nil {
				return err
			}

			return cli.RenderPortfolioSummary(cmd.OutOrStdout(), summary)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")
	return cmd
}

func newPortfolioRiskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "risk",
		Short: "Show basic portfolio concentration and risk summary",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Risk summary:")
			fmt.Println("- No portfolio data available yet")
		},
	}
}

func newPortfolioAddPositionCmd() *cobra.Command {
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

			db, err := store.Open("tick.db")
			if err != nil {
				return err
			}
			defer db.Close()

			repo := store.NewPositionRepository(db)

			position := domain.Position{
				PortfolioName:      portfolioName,
				Ticker:             ticker,
				Quantity:           qty,
				AvgCost:            avgCost,
				InstrumentCurrency: currency,
			}

			if err := repo.Create(context.Background(), position); err != nil {
				return err
			}

			fmt.Printf("Saved %s in portfolio %s: qty=%.4f avg_cost=%.2f %s\n", ticker, portfolioName, qty, avgCost, currency)
			return nil
		},
	}

	cmd.Flags().Float64Var(&qty, "qty", 0, "Position quantity")
	cmd.Flags().Float64Var(&avgCost, "avg-cost", 0, "Average cost basis per unit")
	cmd.Flags().StringVar(&currency, "currency", "USD", "Position currency")
	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")

	return cmd
}

func newPortfolioCreateCmd() *cobra.Command {
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

			db, err := store.Open("tick.db")
			if err != nil {
				return err
			}
			defer db.Close()

			repo := store.NewPortfolioRepository(db)

			p := domain.Portfolio{
				Name:         name,
				BaseCurrency: baseCurrency,
			}

			if err := repo.Create(cmd.Context(), p); err != nil {
				return err
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Portfolio %q saved (base currency: %s)\n",
				name,
				baseCurrency,
			)

			return nil
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
