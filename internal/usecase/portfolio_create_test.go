package usecase

import (
	"context"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase/mocks"
	"go.uber.org/mock/gomock"
)

func TestCreatePortfolioUseCase_Execute_CreatesPortfolioWhenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	portfolios := mocks.NewMockPortfolioRepository(ctrl)

	uc := &CreatePortfolioUseCase{
		portfolios: portfolios,
	}

	portfolios.EXPECT().
		GetByName(gomock.Any(), "main").
		Return(repository.Portfolio{}, domain.ErrPortfolioNotFound)

	portfolios.EXPECT().
		Create(gomock.Any(), repository.Portfolio{
			Name:         "main",
			BaseCurrency: "EUR",
		}).
		Return(nil)

	out, err := uc.Execute(context.Background(), CreatePortfolioUsecaseInput{
		PortfolioName: "main",
		BaseCurrency:  "EUR",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected output, got nil")
	}

	if out.PortfolioName != "main" {
		t.Fatalf("unexpected portfolio name: %q", out.PortfolioName)
	}

	if out.BaseCurrency != "EUR" {
		t.Fatalf("unexpected base currency: %q", out.BaseCurrency)
	}
}

func TestCreatePortfolioUseCase_Execute_ReturnsErrorWhenPortfolioExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	portfolios := mocks.NewMockPortfolioRepository(ctrl)

	uc := &CreatePortfolioUseCase{
		portfolios: portfolios,
	}

	portfolios.EXPECT().
		GetByName(gomock.Any(), "main").
		Return(repository.Portfolio{
			ID:           1,
			Name:         "main",
			BaseCurrency: "EUR",
		}, nil)

	out, err := uc.Execute(context.Background(), CreatePortfolioUsecaseInput{
		PortfolioName: "main",
		BaseCurrency:  "EUR",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}
