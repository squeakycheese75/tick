package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase"
)

type fakeLister struct {
	rows []repository.PortfolioOverview
	err  error
}

func (f fakeLister) List(context.Context) ([]repository.PortfolioOverview, error) {
	return f.rows, f.err
}

func TestListPortfolios_Execute(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	tests := []struct {
		name    string
		lister  fakeLister
		want    []domain.PortfolioListItem
		wantErr error
	}{
		{
			name: "maps every portfolio in order",
			lister: fakeLister{rows: []repository.PortfolioOverview{
				{Name: "main", BaseCurrency: "EUR", PositionCount: 2},
				{Name: "trending", BaseCurrency: "EUR", PositionCount: 5},
			}},
			want: []domain.PortfolioListItem{
				{Name: "main", BaseCurrency: "EUR", PositionCount: 2},
				{Name: "trending", BaseCurrency: "EUR", PositionCount: 5},
			},
		},
		{name: "no portfolios is an empty list, not an error", lister: fakeLister{}, want: []domain.PortfolioListItem{}},
		{name: "store failure is returned", lister: fakeLister{err: boom}, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out, err := usecase.NewListPortfoliosUseCase(tt.lister).Execute(context.Background())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if len(out.Portfolios) != len(tt.want) {
				t.Fatalf("got %+v, want %+v", out.Portfolios, tt.want)
			}
			for i := range tt.want {
				if out.Portfolios[i] != tt.want[i] {
					t.Fatalf("got %+v, want %+v", out.Portfolios, tt.want)
				}
			}
		})
	}
}
