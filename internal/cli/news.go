package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newNewsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "news <ticker>",
		Short: "Show recent news for an asset",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ticker := strings.ToUpper(args[0])
			fmt.Printf("Recent news for %s is not implemented yet\n", ticker)
		},
	}
}
