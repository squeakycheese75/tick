package domain

import (
	"fmt"
	"strings"
	"time"
)

type ImportPortfolioInput struct {
	PortfolioName string                `json:"portfolioName"`
	BaseCurrency  string                `json:"baseCurrency"`
	Positions     []ImportPositionInput `json:"positions"`
}

type ImportPortfolioOutput struct {
	PortfolioName     string
	BaseCurrency      string
	ImportedPositions int
	CreatedPortfolio  bool
}

type ImportPositionInput struct {
	Symbol         string  `json:"symbol"`
	InstrumentType string  `json:"instrumentType"`
	Exchange       string  `json:"exchange"`
	QuoteCurrency  string  `json:"quoteCurrency"`
	Quantity       float64 `json:"quantity"`
	AvgCost        float64 `json:"avgCost"`
}

func (in ImportPortfolioInput) Validate() error {
	if strings.TrimSpace(in.PortfolioName) == "" {
		return Invalid("portfolioName is required")
	}
	if strings.TrimSpace(in.BaseCurrency) == "" {
		return Invalid("baseCurrency is required")
	}
	if len(in.Positions) == 0 {
		return Invalid("at least one position is required")
	}

	for i, p := range in.Positions {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("positions[%d]: %w", i, err)
		}
	}

	return nil
}

func (in ImportPositionInput) Validate() error {
	switch {
	case strings.TrimSpace(in.Symbol) == "":
		return Invalid("symbol is required")
	case strings.TrimSpace(in.InstrumentType) == "":
		return Invalid("instrumentType is required")
	case strings.TrimSpace(in.Exchange) == "":
		return Invalid("exchange is required")
	case strings.TrimSpace(in.QuoteCurrency) == "":
		return Invalid("quoteCurrency is required")
	case in.Quantity <= 0:
		return Invalid("quantity must be greater than 0")
	case in.AvgCost < 0:
		return Invalid("avgCost must be greater than or equal to 0")
	}
	return nil
}

type GetPortfolioSummaryUsecaseInput struct {
	PortfolioName string
}

type PortfoloSummaryReport struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	TotalCost     float64
	TotalPnL      float64
	TotalPnLPct   float64
	// TotalAssetPnL and TotalFXPnL split TotalPnL for the positions whose purchase FX rate is known.
	TotalAssetPnL float64
	TotalFXPnL    float64
	// PricesAsOf is the oldest quote time behind the values; zero when no provider reported one.
	PricesAsOf time.Time
	Positions  []SummaryPosition
}

type GetPortfolioSummaryUsecaseOutput struct {
	Report PortfoloSummaryReport
}

type GetPortfolioRiskInput struct {
	PortfolioName string
}

type GetPortfolioRiskOutput struct {
	PortfolioName     string
	BaseCurrency      string
	PositionCount     int
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type CreatePortfolioUsecaseInput struct {
	PortfolioName string
	BaseCurrency  string
}

func (in CreatePortfolioUsecaseInput) Validate() error {
	in.PortfolioName = strings.TrimSpace(in.PortfolioName)
	in.BaseCurrency = strings.ToUpper(strings.TrimSpace(in.BaseCurrency))

	switch {
	case in.PortfolioName == "":
		return Invalid("portfolio name is required")
	case in.BaseCurrency == "":
		return Invalid("base currency is required")
	}

	return nil
}

type CreatePortfolioUsecaseOutout struct {
	PortfolioName string
	BaseCurrency  string
}

type AddPositionToPortfolioInput struct {
	Symbol         string
	ProviderSymbol string
	InstrumentType string
	Exchange       string
	QuoteCurrency  string
	Qty            float64
	AvgCost        float64
	PortfolioName  string
}

func (i *AddPositionToPortfolioInput) Normalize() {
	i.PortfolioName = strings.TrimSpace(i.PortfolioName)
	i.Symbol = strings.ToUpper(strings.TrimSpace(i.Symbol))
	i.InstrumentType = strings.ToLower(strings.TrimSpace(i.InstrumentType))
	i.Exchange = strings.ToUpper(strings.TrimSpace(i.Exchange))
	i.QuoteCurrency = strings.ToUpper(strings.TrimSpace(i.QuoteCurrency))
}

func (i *AddPositionToPortfolioInput) ApplyDefaults() {
	if i.PortfolioName == "" {
		i.PortfolioName = "main"
	}
}

func (i *AddPositionToPortfolioInput) ValidateBasic() error {
	if i.PortfolioName == "" {
		return Invalid("portfolio name is required")
	}

	if i.Symbol == "" {
		return Invalid("symbol is required")
	}

	if i.Qty <= 0 {
		return Invalid("qty must be greater than 0")
	}

	if i.AvgCost < 0 {
		return Invalid("avg cost must be greater than or equal to 0")
	}

	return nil
}

func (i *AddPositionToPortfolioInput) ValidateResolved() error {
	if i.QuoteCurrency == "" {
		return Invalid("quote currency is required")
	}

	return nil
}

type AddPositionToPortfolioOutput struct {
	Position Position
}

type GetDailyReportInput struct {
	PortfolioName string
	NewsLimit     int
	WithAI        bool
	// Short reports movers and news only, and does not save a snapshot.
	Short bool
	// NoSnapshot reads the report without saving a snapshot of it, so asking has no side effects.
	NoSnapshot bool
}

func (i *GetDailyReportInput) ApplyDefaults() {
	if i.PortfolioName == "" {
		i.PortfolioName = "main"
	}

	if i.NewsLimit <= 0 {
		i.NewsLimit = 2
	}
}

type GetDailyReportOutput struct {
	DailyReport DailyReport
	AISummary   string
}

type SetTargetUseCaseInput struct {
	PortfolioName string
	Symbol        string
	Type          TargetType
	TargetPrice   float64
	QuoteCurrency string
}

type SetTargetUseCaseOutput struct {
	PortfolioName string
	Symbol        string
	Type          TargetType
	TargetPrice   float64
	QuoteCurrency string
	TargetID      int64
}

type ListTargetsUseCaseInput struct {
	PortfolioName string
}

type ListTargetsUseCaseOutput struct {
	PortfolioName string
	Targets       []Target
}

type DeleteTargetUseCaseInput struct {
	TargetID      int64
	PortfolioName string
}

type ConsumePriceUsecaseInput struct {
	Symbol   string    `json:"symbol"`
	Price    float64   `json:"price"`
	Currency string    `json:"currency"`
	AsOf     time.Time `json:"as_of"`
	Source   string    `json:"source"`
}

type ConsumePriceUsecaseOutput struct {
	Symbol   string
	Price    float64
	Currency string
	AsOf     time.Time
	Source   string
}
