package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/buildinfo"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print tick version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "tick version %s\n", buildinfo.Version); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Run `tick config show` to verify your configuration."); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
