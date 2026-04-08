package main

import (
	"fmt"
	"os"

	"github.com/squeakycheese75/tick/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
