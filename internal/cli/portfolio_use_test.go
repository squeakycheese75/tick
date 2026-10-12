package cli

import (
	"strings"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
)

func TestRequirePortfolio(t *testing.T) {
	t.Parallel()

	two := domain.ListPortfoliosOutput{Portfolios: []domain.PortfolioListItem{{Name: "main"}, {Name: "trending"}}}

	tests := []struct {
		name    string
		out     domain.ListPortfoliosOutput
		pick    string
		wantErr string
	}{
		{"exists", two, "trending", ""},
		{"typo lists the real ones", two, "trendng", "choose one of: main, trending"},
		{"nothing created yet", domain.ListPortfoliosOutput{}, "main", "no portfolios yet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := requirePortfolio(tt.out, tt.pick)

			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
