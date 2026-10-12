package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/appdir"
	"github.com/squeakycheese75/tick/internal/cli"
	"github.com/squeakycheese75/tick/internal/domain"
)

func main() {
	rootCmd := cli.NewRootCmd(func() (*app.Runtime, error) {
		dbPath, err := appdir.DBPath()
		if err != nil {
			return nil, err
		}

		return app.BuildRuntime(dbPath)
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)

		if errors.Is(err, domain.ErrPortfolioNotFound) {
			fmt.Fprintln(os.Stderr, "See your portfolios with `tick portfolio list`, or create one with `tick portfolio create <name> --base EUR`.")
		}

		os.Exit(1)
	}
}
