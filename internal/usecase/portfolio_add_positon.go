package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type AddPositionToPortfolioUseCase struct {
	portfolios         PortfolioRepository
	positions          PositionRepository
	instruments        InstrumentRepository
	instrumentResolver InstrumentResolver
}

func NewAddPositionToPortfolioUseCase(
	positionRepo PositionRepository,
	portfolioRepo PortfolioRepository,
	instrumentRepo InstrumentRepository,
	intrumentResolver InstrumentResolver) *AddPositionToPortfolioUseCase {
	return &AddPositionToPortfolioUseCase{
		positions:          positionRepo,
		portfolios:         portfolioRepo,
		instruments:        instrumentRepo,
		instrumentResolver: intrumentResolver,
	}
}

func (uc *AddPositionToPortfolioUseCase) Execute(
	ctx context.Context,
	in domain.AddPositionToPortfolioInput,
) (*domain.AddPositionToPortfolioOutput, error) {
	in.Normalize()
	in.ApplyDefaults()

	if err := in.ValidateBasic(); err != nil {
		return nil, err
	}

	if in.InstrumentType == "" || in.Exchange == "" || in.QuoteCurrency == "" {
		resolved, err := uc.instrumentResolver.Resolve(ctx, in.Symbol)
		if err != nil {
			return nil, fmt.Errorf(
				"could not resolve instrument metadata for %q; please specify --asset-type, --exchange, and --quote-currency",
				in.Symbol,
			)
		}

		if in.InstrumentType == "" {
			in.InstrumentType = resolved.InstrumentType
		}
		if in.Exchange == "" {
			in.Exchange = resolved.Exchange
		}
		if in.QuoteCurrency == "" {
			in.QuoteCurrency = resolved.QuoteCurrency
		}
	}

	if err := in.ValidateResolved(); err != nil {
		return nil, err
	}

	portfolio, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	if err != nil {
		if errors.Is(err, domain.ErrPortfolioNotFound) {
			return nil, fmt.Errorf("portfolio %q not found", in.PortfolioName)
		}
		return nil, fmt.Errorf("get portfolio %q: %w", in.PortfolioName, err)
	}

	instrument, err := uc.instruments.GetOrCreate(ctx, repository.Instrument{
		Symbol:         in.Symbol,
		ProviderSymbol: in.Symbol,
		Exchange:       in.Exchange,
		InstrumentType: in.InstrumentType,
		QuoteCurrency:  in.QuoteCurrency,
	})
	if err != nil {
		return nil, fmt.Errorf("get instrument %q: %w", in.Symbol, err)
	}

	err = uc.positions.Create(ctx, repository.CreatePositionParams{
		InstrumentID: instrument.ID,
		PortfolioID:  portfolio.ID,
		Quantity:     in.Qty,
		AvgCost:      in.AvgCost,
		Currency:     in.QuoteCurrency,
	})
	if err != nil {
		return nil, fmt.Errorf("create position: %w", err)
	}

	return &domain.AddPositionToPortfolioOutput{
		Position: domain.Position{
			PortfolioName: portfolio.Name,
			Instrument: domain.Instrument{
				Symbol:         instrument.Symbol,
				QuoteCurrency:  instrument.QuoteCurrency,
				ProviderSymbol: instrument.ProviderSymbol,
				InstrumentType: domain.InstrumentType(instrument.InstrumentType),
				Exchange:       instrument.Exchange,
			},
			Quantity: in.Qty,
			AvgCost:  in.AvgCost,
		},
	}, nil
}
