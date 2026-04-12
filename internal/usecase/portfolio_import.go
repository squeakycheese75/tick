package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/repository"
)

type ImportPortfolioUseCase struct {
	portfolios  PortfolioRepository
	instruments InstrumentRepository
	positions   PositionRepository
}

func NewImportPortfolioUseCase(
	positionRepo PositionRepository,
	portfolioRepo PortfolioRepository,
	instrumentRepo InstrumentRepository,
) *ImportPortfolioUseCase {
	return &ImportPortfolioUseCase{
		positions:   positionRepo,
		portfolios:  portfolioRepo,
		instruments: instrumentRepo,
	}
}

type ImportPortfolioOutput struct {
	PortfolioName     string
	BaseCurrency      string
	ImportedPositions int
	CreatedPortfolio  bool
}

func (uc *ImportPortfolioUseCase) Execute(
	ctx context.Context,
	in ImportPortfolioInput,
) (*ImportPortfolioOutput, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	portfolio, err := uc.portfolios.GetByName(ctx, in.PortfolioName)
	createdPortfolio := false

	if err != nil {
		if !errors.Is(err, domain.ErrPortfolioNotFound) {
			return nil, fmt.Errorf("get portfolio %q: %w", in.PortfolioName, err)
		}

		if err := uc.portfolios.Create(ctx, repository.Portfolio{
			Name:         in.PortfolioName,
			BaseCurrency: strings.ToUpper(strings.TrimSpace(in.BaseCurrency)),
		}); err != nil {
			return nil, fmt.Errorf("create portfolio %q: %w", in.PortfolioName, err)
		}

		createdPortfolio = true

		portfolio, err = uc.portfolios.GetByName(ctx, in.PortfolioName)
		if err != nil {
			return nil, fmt.Errorf("reload portfolio %q: %w", in.PortfolioName, err)
		}
	}

	for _, pos := range in.Positions {
		instrument, err := uc.instruments.GetOrCreate(ctx, repository.Instrument{
			Symbol:         strings.ToUpper(strings.TrimSpace(pos.Symbol)),
			ProviderSymbol: strings.ToUpper(strings.TrimSpace(pos.Symbol)),
			AssetType:      strings.ToLower(strings.TrimSpace(pos.AssetType)),
			Exchange:       strings.ToUpper(strings.TrimSpace(pos.Exchange)),
			QuoteCurrency:  strings.ToUpper(strings.TrimSpace(pos.QuoteCurrency)),
		})
		if err != nil {
			return nil, fmt.Errorf(
				"get or create instrument %q on exchange %q: %w",
				pos.Symbol,
				pos.Exchange,
				err,
			)
		}

		err = uc.positions.Create(ctx, repository.CreatePositionParams{
			InstrumentID: instrument.ID,
			PortfolioID:  portfolio.ID,
			Quantity:     pos.Quantity,
			AvgCost:      pos.AvgCost,
			Currency:     strings.ToUpper(strings.TrimSpace(pos.QuoteCurrency)),
		})
		if err != nil {
			return nil, fmt.Errorf(
				"create position for symbol %q in portfolio %q: %w",
				pos.Symbol,
				portfolio.Name,
				err,
			)
		}
	}

	return &ImportPortfolioOutput{
		PortfolioName:     portfolio.Name,
		BaseCurrency:      portfolio.BaseCurrency,
		ImportedPositions: len(in.Positions),
		CreatedPortfolio:  createdPortfolio,
	}, nil
}
