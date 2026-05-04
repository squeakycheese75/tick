package domain

import (
	"fmt"
	"strings"
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
		return fmt.Errorf("portfolioName is required")
	}
	if strings.TrimSpace(in.BaseCurrency) == "" {
		return fmt.Errorf("baseCurrency is required")
	}
	if len(in.Positions) == 0 {
		return fmt.Errorf("at least one position is required")
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
		return fmt.Errorf("symbol is required")
	case strings.TrimSpace(in.InstrumentType) == "":
		return fmt.Errorf("instrumentType is required")
	case strings.TrimSpace(in.Exchange) == "":
		return fmt.Errorf("exchange is required")
	case strings.TrimSpace(in.QuoteCurrency) == "":
		return fmt.Errorf("quoteCurrency is required")
	case in.Quantity <= 0:
		return fmt.Errorf("quantity must be greater than 0")
	case in.AvgCost < 0:
		return fmt.Errorf("avgCost must be greater than or equal to 0")
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
	Positions     []SummaryPosition
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
		return fmt.Errorf("portfolio name is required")
	case in.BaseCurrency == "":
		return fmt.Errorf("base currency is required")
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
		return fmt.Errorf("portfolio name is required")
	}

	if i.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if i.Qty <= 0 {
		return fmt.Errorf("qty must be greater than 0")
	}

	if i.AvgCost < 0 {
		return fmt.Errorf("avg cost must be greater than or equal to 0")
	}

	return nil
}

func (i *AddPositionToPortfolioInput) ValidateResolved() error {
	if i.QuoteCurrency == "" {
		return fmt.Errorf("quote currency is required")
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

type GetMorningBriefUsecaseInput struct {
	PortfolioName string
	NewsLimit     int
}

func (i *GetMorningBriefUsecaseInput) ApplyDefaults() {
	if i.PortfolioName == "" {
		i.PortfolioName = "main"
	}
}

type GetMorningBriefUsecaseOutput struct {
	Report BriefReport
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
