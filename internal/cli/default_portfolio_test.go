package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// These tests change HOME and the environment, so they cannot run in parallel.

func TestResolvePortfolio_Precedence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(portfolioEnvVar, "")

	if got := resolvePortfolio(""); got != "main" {
		t.Fatalf("nothing set: got %q, want main", got)
	}

	if err := writeDefaultPortfolio("trending"); err != nil {
		t.Fatalf("write default: %v", err)
	}
	if got := resolvePortfolio(""); got != "trending" {
		t.Fatalf("stored default: got %q, want trending", got)
	}

	t.Setenv(portfolioEnvVar, " pension ")
	if got := resolvePortfolio(""); got != "pension" {
		t.Fatalf("env beats stored default: got %q, want pension", got)
	}

	if got := resolvePortfolio("explicit"); got != "explicit" {
		t.Fatalf("flag beats everything: got %q, want explicit", got)
	}
}

func TestReadDefaultPortfolio_BlankFileMeansNone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(portfolioEnvVar, "")

	if err := writeDefaultPortfolio("  "); err != nil {
		t.Fatalf("write default: %v", err)
	}
	if got := readDefaultPortfolio(); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestAddPortfolioFlag_ResolvesBeforeTheCommandRuns(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(portfolioEnvVar, "")

	if err := writeDefaultPortfolio("trending"); err != nil {
		t.Fatalf("write default: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"uses the stored default", nil, "trending"},
		{"flag wins", []string{"--portfolio", "other"}, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var name, seen string
			previousRan := false

			cmd := &cobra.Command{
				Use:     "x",
				PreRunE: func(*cobra.Command, []string) error { previousRan = true; return nil },
				RunE:    func(*cobra.Command, []string) error { seen = name; return nil },
			}
			addPortfolioFlag(cmd, &name)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute: %v", err)
			}
			if seen != tt.want {
				t.Fatalf("portfolio = %q, want %q", seen, tt.want)
			}
			if !previousRan {
				t.Fatal("an existing PreRunE must still run")
			}
		})
	}
}

func TestAddPortfolioFlag_ErrorFromExistingPreRunStopsTheCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var name string
	ran := false

	cmd := &cobra.Command{
		Use:     "x",
		PreRunE: func(*cobra.Command, []string) error { return errBoom },
		RunE:    func(*cobra.Command, []string) error { ran = true; return nil },
	}
	addPortfolioFlag(cmd, &name)

	if err := cmd.Execute(); err == nil || ran {
		t.Fatalf("expected the PreRunE error to stop the command, err=%v ran=%v", err, ran)
	}
}

type boomError struct{}

func (boomError) Error() string { return "boom" }

var errBoom = boomError{}
