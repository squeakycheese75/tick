package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/squeakycheese75/tick/internal/domain"
)

func newPricesCommand(runtimeBuilder RuntimeBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prices",
		Short: "Manage consumed prices",
	}

	cmd.AddCommand(newPricesConsumeCmd(runtimeBuilder))

	return cmd
}

type consumedPriceInput struct {
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	AsOf     string  `json:"as_of"`
	Source   string  `json:"source"`
}

func newPricesConsumeCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "consume --file <file>",
		Short: "Consume prices from a JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			b, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read prices file: %w", err)
			}

			var prices []consumedPriceInput
			if err := json.Unmarshal(b, &prices); err != nil {
				return fmt.Errorf("unmarshal prices file: %w", err)
			}

			if len(prices) == 0 {
				return fmt.Errorf("no prices to consume")
			}

			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			for _, p := range prices {
				asOf, err := parseAsOf(p.AsOf)
				if err != nil {
					return fmt.Errorf("parse as_of for %s: %w", p.Symbol, err)
				}

				if err := app.ConsumePrices.Execute(ctx, domain.ConsumePriceUsecaseInput{
					Symbol:   p.Symbol,
					Price:    p.Price,
					Currency: p.Currency,
					AsOf:     asOf,
					Source:   p.Source,
				}); err != nil {
					return fmt.Errorf("consume price for %s: %w", p.Symbol, err)
				}

				fmt.Printf(
					"Consumed %s %.4f %s (%s)\n",
					p.Symbol,
					p.Price,
					p.Currency,
					asOf,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to prices JSON file")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func parseAsOf(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}

	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	return time.Parse(time.RFC3339, s)
}
