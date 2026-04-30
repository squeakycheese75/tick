package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/render"
)

func newTargetCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target",
		Short: "Manage portfolio price targets",
	}

	cmd.AddCommand(newTargetSetCmd(runtimeBuilder))
	cmd.AddCommand(newTargetListCmd(runtimeBuilder))
	// cmd.AddCommand(newTargetDeleteCmd(runtimeBuilder))

	return cmd
}

func newTargetSetCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var portfolioName string
	var takeProfit float64
	var stopLoss float64
	var quoteCurrency string

	cmd := &cobra.Command{
		Use:   "set SYMBOL",
		Short: "Set a portfolio price target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if takeProfit == 0 && stopLoss == 0 {
				return fmt.Errorf("provide either --take-profit or --stop-loss")
			}

			if takeProfit > 0 && stopLoss > 0 {
				return fmt.Errorf("provide only one of --take-profit or --stop-loss")
			}

			targetType := domain.TargetTypeTakeProfit
			targetPrice := takeProfit

			if stopLoss > 0 {
				targetType = domain.TargetTypeStopLoss
				targetPrice = stopLoss
			}

			rt, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := rt.SetTarget.Execute(cmd.Context(), domain.SetTargetUseCaseInput{
				PortfolioName: portfolioName,
				Symbol:        strings.ToUpper(args[0]),
				Type:          targetType,
				TargetPrice:   targetPrice,
				QuoteCurrency: strings.ToUpper(quoteCurrency),
			})
			if err != nil {
				return err
			}

			return render.RenderSetTarget(cmd.OutOrStdout(), *out)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")
	cmd.Flags().Float64Var(&takeProfit, "take-profit", 0, "Take-profit price")
	cmd.Flags().Float64Var(&stopLoss, "stop-loss", 0, "Stop-loss price")
	cmd.Flags().StringVar(&quoteCurrency, "currency", "USD", "Quote currency")

	return cmd
}

func newTargetListCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var portfolioName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List portfolio targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeBuilder()
			if err != nil {
				return err
			}

			out, err := rt.ListTargets.Execute(cmd.Context(), domain.ListTargetsUseCaseInput{
				PortfolioName: portfolioName,
			})
			if err != nil {
				return err
			}

			return render.RenderListTargets(cmd.OutOrStdout(), out)
		},
	}

	cmd.Flags().StringVar(&portfolioName, "portfolio", "main", "Portfolio name")

	return cmd
}
