package cli

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

func newShellCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Start interactive shell",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShell(rootCmd)
		},
	}
}

func runShell(rootCmd *cobra.Command) error {
	rl, err := readline.New("tick> ")
	if err != nil {
		return fmt.Errorf("create shell: %w", err)
	}
	defer func() {
		_ = rl.Close()
	}()

	fmt.Println("tick interactive mode")
	fmt.Println("Type 'help' for commands, 'exit' or 'quit' to leave.")
	fmt.Println("")

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			return nil
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch strings.ToLower(line) {
		case "exit", "quit":
			return nil
		}

		args := strings.Fields(line)

		rootCmd.SetArgs(args)
		err = rootCmd.Execute()
		if err != nil {
			fmt.Println("Error:", err)
		}

		fmt.Println("")
	}
}
