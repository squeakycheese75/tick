package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <ticker>",
		Short: "Show a quick overview for an asset",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ticker := strings.ToUpper(args[0])
			fmt.Printf("%s — asset overview not implemented yet\n", ticker)
		},
	}
}
