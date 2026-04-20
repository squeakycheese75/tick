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
	in AddPositionToPortfolioInput,
) (*AddPositionToPortfolioOutput, error) {
	in.Normalize()
	in.ApplyDefaults()

	if err := in.ValidateBasic(); err != nil {
		return nil, err
	}

	if in.AssetType == "" || in.Exchange == "" || in.QuoteCurrency == "" {
		resolved, err := uc.instrumentResolver.Resolve(ctx, in.Symbol)
		if err != nil {
			return nil, fmt.Errorf(
				"could not resolve instrument metadata for %q; please specify --asset-type, --exchange, and --quote-currency",
				in.Symbol,
			)
		}

		if in.AssetType == "" {
			in.AssetType = resolved.AssetType
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
		AssetType:      in.AssetType,
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
		if errors.Is(err, domain.ErrPortfolioAlreadyExists) {
			return nil, fmt.Errorf(
				"position for %q already exists in portfolio %q",
				instrument.Symbol,
				portfolio.Name,
			)
		}
		return nil, fmt.Errorf("create position: %w", err)
	}

	return &AddPositionToPortfolioOutput{
		PortfolioName: portfolio.Name,
		Symbol:        instrument.Symbol,
		Qty:           in.Qty,
		AvgCost:       in.AvgCost,
		QuoteCurrency: in.QuoteCurrency,
	}, nil
}
