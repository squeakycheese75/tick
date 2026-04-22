package usecase

import (
	"context"
	"testing"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
	"github.com/squeakycheese75/tick/internal/usecase/mocks"
	"go.uber.org/mock/gomock"
)

func TestAddPositionToPortfolioUseCase_Execute_CreatesPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	portfolios := mocks.NewMockPortfolioRepository(ctrl)
	instruments := mocks.NewMockInstrumentRepository(ctrl)
	positions := mocks.NewMockPositionRepository(ctrl)
	instrumentResolver := mocks.NewMockInstrumentResolver(ctrl)

	uc := NewAddPositionToPortfolioUseCase(positions, portfolios, instruments, instrumentResolver)

	in := domain.AddPositionToPortfolioInput{
		PortfolioName:  "main",
		Symbol:         "NVDA",
		InstrumentType: "equity",
		Exchange:       "NASDAQ",
		QuoteCurrency:  "USD",
		Qty:            10,
		AvgCost:        400,
	}

	portfolios.EXPECT().
		GetByName(gomock.Any(), "main").
		Return(repository.Portfolio{
			ID:           1,
			Name:         "main",
			BaseCurrency: "EUR",
		}, nil)

	instruments.EXPECT().
		GetOrCreate(gomock.Any(), repository.Instrument{
			Symbol:         "NVDA",
			ProviderSymbol: "NVDA",
			InstrumentType: "equity",
			Exchange:       "NASDAQ",
			QuoteCurrency:  "USD",
		}).
		Return(repository.Instrument{
			ID:             42,
			Symbol:         "NVDA",
			ProviderSymbol: "NVDA",
			InstrumentType: "equity",
			Exchange:       "NASDAQ",
			QuoteCurrency:  "USD",
		}, nil)

	positions.EXPECT().
		Create(gomock.Any(), repository.CreatePositionParams{
			InstrumentID: 42,
			PortfolioID:  1,
			Quantity:     10,
			AvgCost:      400,
			Currency:     "USD",
		}).
		Return(nil)

	out, err := uc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected output, got nil")
	}

	if out.PortfolioName != "main" {
		t.Fatalf("unexpected portfolio name: %q", out.PortfolioName)
	}

	if out.Symbol != "NVDA" {
		t.Fatalf("unexpected symbol: %q", out.Symbol)
	}
}

func TestAddPositionToPortfolioUseCase_Execute_ReturnsErrorWhenPositionExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	portfolios := mocks.NewMockPortfolioRepository(ctrl)
	instruments := mocks.NewMockInstrumentRepository(ctrl)
	positions := mocks.NewMockPositionRepository(ctrl)
	instrumentResolver := mocks.NewMockInstrumentResolver(ctrl)

	uc := NewAddPositionToPortfolioUseCase(positions, portfolios, instruments, instrumentResolver)

	in := domain.AddPositionToPortfolioInput{
		PortfolioName:  "main",
		Symbol:         "NVDA",
		InstrumentType: "equity",
		Exchange:       "NASDAQ",
		QuoteCurrency:  "USD",
		Qty:            10,
		AvgCost:        400,
	}

	portfolios.EXPECT().
		GetByName(gomock.Any(), "main").
		Return(repository.Portfolio{
			ID:   1,
			Name: "main",
		}, nil)

	instruments.EXPECT().
		GetOrCreate(gomock.Any(), gomock.Any()).
		Return(repository.Instrument{
			ID:     42,
			Symbol: "NVDA",
		}, nil)

	positions.EXPECT().
		Create(gomock.Any(), repository.CreatePositionParams{
			InstrumentID: 42,
			PortfolioID:  1,
			Quantity:     10,
			AvgCost:      400,
			Currency:     "USD",
		}).
		Return(domain.ErrPositionAlreadyExists)

	out, err := uc.Execute(context.Background(), in)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}
