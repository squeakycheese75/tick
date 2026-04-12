package usecase

import (
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/report"
)

type GetPortfolioSummaryUsecaseInput struct {
	PortfolioName string
}

type GetPortfolioSummaryUsecaseOutput struct {
	PortfolioName string
	BaseCurrency  string
	TotalValue    float64
	Positions     []SummaryPosition
}

type SummaryPosition struct {
	Symbol             string
	BaseCurrency       string
	InstrumentCurrency string
	AvgCost            float64
	ConvertedPrice     float64
	QuotedPrice        float64
	Weight             float64
	MarketValueBase    float64
	FXRate             float64
	Quantity           float64
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
	AssetType      string
	Exchange       string
	QuoteCurrency  string
	Qty            float64
	AvgCost        float64
	PortfolioName  string
}

func (i *AddPositionToPortfolioInput) Validate() error {
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

	if i.QuoteCurrency == "" {
		return fmt.Errorf("quote currency is required")
	}

	return nil
}

type AddPositionToPortfolioOutput struct {
	Symbol         string
	ProviderSymbol string
	AssetType      string
	Exchange       string
	Qty            float64
	QuoteCurrency  string
	AvgCost        float64
	PortfolioName  string
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

type GetDailyReportInput struct {
	PortfolioName string
	NewsLimit     int
	WithAI        bool
}

type GetDailyReportOutput struct {
	DailyReport report.DailyReport
	AISummary   string
}

type TickerNews struct {
	Ticker    string
	Headlines []NewsHeadline
}

type DailyHolding struct {
	Ticker          string
	Weight          float64
	MarketValueBase float64
	QuotedPrice     float64
	PriceCurrency   string
	ChangePercent   float64
}

type DailyRisk struct {
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

type DailyNews struct {
	Ticker    string
	Headlines []NewsHeadline
}

type NewsHeadline struct {
	Title string
	URL   string
}

type ImportPortfolioInput struct {
	PortfolioName string                `json:"portfolioName"`
	BaseCurrency  string                `json:"baseCurrency"`
	Positions     []ImportPositionInput `json:"positions"`
}

type ImportPositionInput struct {
	Symbol        string  `json:"symbol"`
	AssetType     string  `json:"assetType"`
	Exchange      string  `json:"exchange"`
	QuoteCurrency string  `json:"quoteCurrency"`
	Quantity      float64 `json:"quantity"`
	AvgCost       float64 `json:"avgCost"`
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
	case strings.TrimSpace(in.AssetType) == "":
		return fmt.Errorf("assetType is required")
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
