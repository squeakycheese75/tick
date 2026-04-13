package main

import (
	"fmt"
	"os"

	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCmd(func() (*app.Runtime, error) {
		dbPath, err := app.DefaultDBPath()
		if err != nil {
			return nil, err
		}

		return app.BuildRuntime(dbPath)
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
