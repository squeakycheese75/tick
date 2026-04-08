package main

import (
	"fmt"
	"os"

	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/cli"
)

func main() {
	app, err := app.BuildRuntime("tick.db")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rootCmd := cli.NewRootCmd(app)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
