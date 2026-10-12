package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase"
)

type fakePortfolios struct {
	getErr    error
	createErr error
	created   []repository.Portfolio
}

func (f *fakePortfolios) GetByName(context.Context, string) (repository.Portfolio, error) {
	return repository.Portfolio{}, f.getErr
}

func (f *fakePortfolios) Create(_ context.Context, p repository.Portfolio) error {
	f.created = append(f.created, p)
	return f.createErr
}

func TestCreatePortfolio_Execute_ErrorsAreClassifiedForCallers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		repo  *fakePortfolios
		in    domain.CreatePortfolioUsecaseInput
		want  domain.ErrorKind
		wrote bool
	}{
		{"creates a new one", &fakePortfolios{getErr: domain.ErrPortfolioNotFound}, domain.CreatePortfolioUsecaseInput{PortfolioName: "p", BaseCurrency: "EUR"}, -1, true},
		{"already there", &fakePortfolios{}, domain.CreatePortfolioUsecaseInput{PortfolioName: "p", BaseCurrency: "EUR"}, domain.KindConflict, false},
		{"missing name", &fakePortfolios{}, domain.CreatePortfolioUsecaseInput{BaseCurrency: "EUR"}, domain.KindInvalid, false},
		{"lookup fails", &fakePortfolios{getErr: errors.New("boom")}, domain.CreatePortfolioUsecaseInput{PortfolioName: "p", BaseCurrency: "EUR"}, domain.KindInternal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := usecase.NewCreatePortfolioUseCase(tt.repo).Execute(context.Background(), tt.in)

			if tt.want == -1 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err == nil || domain.KindOf(err) != tt.want {
				t.Fatalf("kind = %s (err: %v), want %s", domain.KindOf(err), err, tt.want)
			}

			if (len(tt.repo.created) == 1) != tt.wrote {
				t.Fatalf("created = %v, want written=%v", tt.repo.created, tt.wrote)
			}
		})
	}
}
