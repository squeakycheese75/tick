package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/appdir"
	"github.com/squeakycheese75/tick/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration commands",
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigInitCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show effective configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.LoadConfig()
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), cfg.String())

			if err := cfg.Validate(); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "\nWarning: %v\n", err)
			}

			return nil
		},
	}
}

func newConfigInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create default config file (~/.tick/config.env)",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := appdir.ConfigPath()
			if err != nil {
				return err
			}

			if _, err := os.Stat(configPath); err == nil && !force {
				return fmt.Errorf(
					"config already exists at %s (use --force to overwrite)",
					configPath,
				)
			}

			if err := os.WriteFile(configPath, []byte(config.DefaultEnv), 0o644); err != nil {
				return fmt.Errorf("write config file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created config at %s\n", configPath)
			fmt.Fprintln(cmd.OutOrStdout(), "Edit this file to configure tick.")
			fmt.Fprintln(cmd.OutOrStdout(), "Run `tick config show` to verify your configuration.")

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config")

	return cmd
}
