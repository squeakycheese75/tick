package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/buildinfo"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print tick version information",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "tick version %s\n", buildinfo.Version)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "commit: %s\n", buildinfo.Commit)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "built: %s\n", buildinfo.BuildDate)
		},
	}
}
