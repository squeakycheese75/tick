package api

import (
	"fmt"
	"time"

	"github.com/squeakycheese75/tick/internal/domain"
)

// Ratios, such as pnlRatio and weight, are fractions: 0.25 is 25%. Money is in the currency
// the field's name or its parent says.

type createPortfolioRequest struct {
	Name         string `json:"name"`
	BaseCurrency string `json:"baseCurrency"`
}

type portfolioResponse struct {
	Name         string `json:"name"`
	BaseCurrency string `json:"baseCurrency"`
}

type portfolioListResponse struct {
	Portfolios []portfolioListItem `json:"portfolios"`
}

type portfolioListItem struct {
	Name          string `json:"name"`
	BaseCurrency  string `json:"baseCurrency"`
	PositionCount int    `json:"positionCount"`
}

type tradeRequest struct {
	Ref            string  `json:"ref"`
	Date           string  `json:"date"`
	Symbol         string  `json:"symbol"`
	Side           string  `json:"side"`
	Quantity       float64 `json:"quantity"`
	Price          float64 `json:"price"`
	Currency       string  `json:"currency"`
	Fee            float64 `json:"fee"`
	FXRate         float64 `json:"fxRate"`
	Exchange       string  `json:"exchange"`
	InstrumentType string  `json:"instrumentType"`
	QuoteCurrency  string  `json:"quoteCurrency"`
}

type importTradesRequest struct {
	Trades []tradeRequest `json:"trades"`
}

type importTradesResponse struct {
	Portfolio string `json:"portfolio"`
	Applied   int    `json:"applied"`
	Skipped   int    `json:"skipped"`
}

var tradeDateLayouts = []string{"2006-01-02", time.RFC3339}

func toTradeInputs(reqs []tradeRequest) ([]domain.TradeInput, error) {
	trades := make([]domain.TradeInput, 0, len(reqs))

	for i, r := range reqs {
		date, err := parseTradeDate(r.Date)
		if err != nil {
			return nil, domain.Invalid("trades[%d]: %s", i, err)
		}

		trades = append(trades, domain.TradeInput{
			Ref:            r.Ref,
			Date:           date,
			Symbol:         r.Symbol,
			Side:           domain.TradeSide(r.Side),
			Quantity:       r.Quantity,
			Price:          r.Price,
			Currency:       r.Currency,
			Fee:            r.Fee,
			FXRate:         r.FXRate,
			Exchange:       r.Exchange,
			InstrumentType: r.InstrumentType,
			QuoteCurrency:  r.QuoteCurrency,
		})
	}

	return trades, nil
}

func parseTradeDate(value string) (time.Time, error) {
	for _, layout := range tradeDateLayouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", value)
}

type summaryResponse struct {
	Portfolio     string            `json:"portfolio"`
	BaseCurrency  string            `json:"baseCurrency"`
	PricesAsOf    *time.Time        `json:"pricesAsOf,omitempty"`
	TotalValue    float64           `json:"totalValue"`
	TotalCost     float64           `json:"totalCost"`
	TotalPnL      float64           `json:"totalPnL"`
	TotalPnLRatio float64           `json:"totalPnLRatio"`
	TotalAssetPnL float64           `json:"totalAssetPnL"`
	TotalFXPnL    float64           `json:"totalFXPnL"`
	Positions     []summaryPosition `json:"positions"`
}

type summaryPosition struct {
	Symbol        string  `json:"symbol"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	PriceCurrency string  `json:"priceCurrency"`
	Value         float64 `json:"value"`
	Cost          float64 `json:"cost"`
	PnL           float64 `json:"pnl"`
	PnLRatio      float64 `json:"pnlRatio"`
	// AssetPnL and FXPnL are null when the purchase FX rate was not recorded.
	AssetPnL *float64 `json:"assetPnL"`
	FXPnL    *float64 `json:"fxPnL"`
	Weight   float64  `json:"weight"`
}

func toSummaryResponse(r domain.PortfoloSummaryReport) summaryResponse {
	positions := make([]summaryPosition, 0, len(r.Positions))

	for _, p := range r.Positions {
		sp := summaryPosition{
			Symbol:        p.Symbol,
			Quantity:      p.Quantity,
			Price:         p.QuotedPrice,
			PriceCurrency: p.InstrumentCurrency,
			Value:         p.MarketValueBase,
			Cost:          p.CostBasisBase,
			PnL:           p.UnrealizedPnL,
			PnLRatio:      p.UnrealizedPnLPct,
			Weight:        p.Weight,
		}
		if p.FXKnown {
			asset, fx := p.AssetPnL, p.FXPnL
			sp.AssetPnL, sp.FXPnL = &asset, &fx
		}

		positions = append(positions, sp)
	}

	return summaryResponse{
		Portfolio:     r.PortfolioName,
		BaseCurrency:  r.BaseCurrency,
		PricesAsOf:    timePtr(r.PricesAsOf),
		TotalValue:    r.TotalValue,
		TotalCost:     r.TotalCost,
		TotalPnL:      r.TotalPnL,
		TotalPnLRatio: r.TotalPnLPct,
		TotalAssetPnL: r.TotalAssetPnL,
		TotalFXPnL:    r.TotalFXPnL,
		Positions:     positions,
	}
}

type dailyResponse struct {
	Portfolio       dailyPortfolio     `json:"portfolio"`
	Holdings        []dailyHolding     `json:"holdings"`
	Risk            *dailyRisk         `json:"risk,omitempty"`
	News            []dailyNews        `json:"news"`
	Attention       []string           `json:"attention"`
	Targets         []dailyTarget      `json:"targets"`
	ValuationIssues []valuationIssueDT `json:"valuationIssues"`
	AISummary       string             `json:"aiSummary,omitempty"`
}

type dailyPortfolio struct {
	Name         string     `json:"name"`
	BaseCurrency string     `json:"baseCurrency"`
	TotalValue   float64    `json:"totalValue"`
	PricesAsOf   *time.Time `json:"pricesAsOf,omitempty"`
}

type dailyHolding struct {
	Symbol        string  `json:"symbol"`
	Quantity      float64 `json:"quantity"`
	Weight        float64 `json:"weight"`
	Value         float64 `json:"value"`
	Price         float64 `json:"price"`
	PriceCurrency string  `json:"priceCurrency"`
	// ChangeRatio is the price move since the previous close; it may be the last session's.
	ChangeRatio float64 `json:"changeRatio"`
	ChangeValue float64 `json:"changeValue"`
}

type dailyRisk struct {
	LargestPosition   string   `json:"largestPosition"`
	LargestWeight     float64  `json:"largestWeight"`
	Top3Concentration float64  `json:"top3Concentration"`
	Observations      []string `json:"observations"`
}

type dailyNews struct {
	Ticker    string          `json:"ticker"`
	Headlines []newsHeadlineD `json:"headlines"`
}

type newsHeadlineD struct {
	Title string `json:"title"`
	URL   string `json:"url,omitempty"`
}

type dailyTarget struct {
	Symbol        string  `json:"symbol"`
	Type          string  `json:"type"`
	CurrentPrice  float64 `json:"currentPrice"`
	TargetPrice   float64 `json:"targetPrice"`
	Currency      string  `json:"currency"`
	Hit           bool    `json:"hit"`
	DistanceRatio float64 `json:"distanceRatio"`
}

type valuationIssueDT struct {
	Symbol   string  `json:"symbol"`
	Type     string  `json:"type"`
	Message  string  `json:"message"`
	Hint     string  `json:"hint,omitempty"`
	Quantity float64 `json:"quantity"`
}

func toDailyResponse(out domain.GetDailyReportOutput) dailyResponse {
	r := out.DailyReport

	resp := dailyResponse{
		Portfolio: dailyPortfolio{
			Name:         r.Portfolio.Name,
			BaseCurrency: r.Portfolio.BaseCurrency,
			TotalValue:   r.Portfolio.TotalValue,
			PricesAsOf:   timePtr(r.Portfolio.PricesAsOf),
		},
		Holdings:        make([]dailyHolding, 0, len(r.TopHoldings.Holdings)),
		News:            make([]dailyNews, 0, len(r.News)),
		Attention:       append([]string{}, r.Attention...),
		Targets:         make([]dailyTarget, 0, len(r.Targets)),
		ValuationIssues: make([]valuationIssueDT, 0, len(r.ValuationIssues)),
		AISummary:       out.AISummary,
	}

	for _, h := range r.TopHoldings.Holdings {
		resp.Holdings = append(resp.Holdings, dailyHolding{
			Symbol:        h.Symbol,
			Quantity:      h.Quantity,
			Weight:        h.Weight,
			Value:         h.MarketValueBase,
			Price:         h.QuotedPrice,
			PriceCurrency: h.PriceCurrency,
			ChangeRatio:   h.ChangePercent / 100,
			ChangeValue:   h.ChangeAbsolute,
		})
	}

	if r.Risk.LargestPosition != "" {
		resp.Risk = &dailyRisk{
			LargestPosition:   r.Risk.LargestPosition,
			LargestWeight:     r.Risk.LargestWeight,
			Top3Concentration: r.Risk.Top3Concentration,
			Observations:      append([]string{}, r.Risk.Observations...),
		}
	}

	for _, n := range r.News {
		headlines := make([]newsHeadlineD, 0, len(n.Headlines))
		for _, h := range n.Headlines {
			headlines = append(headlines, newsHeadlineD{Title: h.Title, URL: h.URL})
		}

		resp.News = append(resp.News, dailyNews{Ticker: n.Ticker, Headlines: headlines})
	}

	for _, t := range r.Targets {
		resp.Targets = append(resp.Targets, dailyTarget{
			Symbol:        t.Symbol,
			Type:          string(t.Type),
			CurrentPrice:  t.CurrentPrice,
			TargetPrice:   t.TargetPrice,
			Currency:      t.Currency,
			Hit:           t.Hit,
			DistanceRatio: t.DistancePct,
		})
	}

	for _, v := range r.ValuationIssues {
		resp.ValuationIssues = append(resp.ValuationIssues, valuationIssueDT{
			Symbol: v.Symbol, Type: string(v.Type), Message: v.Message, Hint: v.Hint, Quantity: v.Quantity,
		})
	}

	return resp
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}

	utc := t.UTC()

	return &utc
}
