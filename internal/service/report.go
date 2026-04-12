package service

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/report"
)

type ReportService struct {
	portfolioSvc *PortfolioService
	newsSvc      *NewsService
	insights     *PortfolioInsights
}

func NewReportService(portfolioSvc *PortfolioService, newsSvc *NewsService, insights *PortfolioInsights) *ReportService {
	return &ReportService{
		portfolioSvc: portfolioSvc,
		newsSvc:      newsSvc,
		insights:     insights,
	}
}

func (s *ReportService) BuildDailyReport(
	ctx context.Context,
	portfolioName string,
	newsLimit int,
) (report.DailyReport, error) {
	portfolioAnalysis, err := s.portfolioSvc.GetAnalysis(ctx, portfolioName)
	if err != nil {
		return report.DailyReport{}, fmt.Errorf("get portfolio analysis: %w", err)
	}

	portfolioRisk, err := s.portfolioSvc.GetRisk(ctx, portfolioName)
	if err != nil {
		return report.DailyReport{}, fmt.Errorf("get portfolio risk: %w", err)
	}

	topPositions := s.insights.TopHoldings(portfolioAnalysis, 3)
	topHoldings := make([]report.TopHoldingReport, 0, len(topPositions))
	for _, pos := range topPositions {
		topHoldings = append(topHoldings, report.TopHoldingReport{
			Symbol:          pos.Symbol,
			Weight:          pos.Weight,
			MarketValueBase: pos.MarketValueBase,
			QuotedPrice:     pos.QuotedPrice,
			PriceCurrency:   pos.PriceCurrency,
			ChangePercent:   pos.QuotedChangePct,
		})
	}

	out := report.DailyReport{
		PortfolioName: portfolioAnalysis.PortfolioName,
		BaseCurrency:  portfolioAnalysis.BaseCurrency,
		TotalValue:    portfolioAnalysis.TotalValue,
		TopHoldings:   topHoldings,
		Risk: report.RiskReport{
			LargestPosition:   portfolioRisk.LargestPosition,
			LargestWeight:     portfolioRisk.LargestWeight,
			Top3Concentration: portfolioRisk.Top3Concentration,
			Observations:      append([]string(nil), portfolioRisk.Observations...),
		},
		Attention: s.insights.AttentionSignals(portfolioAnalysis, portfolioRisk),
		News:      make([]report.TickerNewsReport, 0, len(topHoldings)),
	}

	for _, holding := range out.TopHoldings {
		tickerNews, err := s.newsSvc.GetNews(ctx, holding.Symbol, newsLimit)
		if err != nil {
			return report.DailyReport{}, fmt.Errorf("get news for %s: %w", holding.Symbol, err)
		}

		out.News = append(out.News, tickerNews)
	}

	return out, nil
}
