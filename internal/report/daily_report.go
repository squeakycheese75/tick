package report

import (
	"context"
	"fmt"
	"sync"

	"github.com/squeakycheese75/tick/internal/domain/analysis"
)

type (
	PortfolioService interface {
		GetAnalysis(ctx context.Context, portfolioName string) (analysis.PortfolioAnalysis, error)
		GetRisk(ctx context.Context, portfolioName string) (analysis.PortfolioRisk, error)
	}
	NewsService interface {
		GetNews(ctx context.Context, ticker string, newsLimit int) (TickerNewsReport, error)
	}
	PortfolioInsights interface {
		TopHoldings(portfolioAnalysis analysis.PortfolioAnalysis, limit int) []analysis.AnalyzedPosition
		AttentionSignals(portfolioAnalysis analysis.PortfolioAnalysis, portfolioRisk analysis.PortfolioRisk) []string
	}
)

type ReportBuilder struct {
	portfolioSvc PortfolioService
	newsSvc      NewsService
	insights     PortfolioInsights
}

func NewReportBuilder(portfolioSvc PortfolioService, newsSvc NewsService, insights PortfolioInsights) *ReportBuilder {
	return &ReportBuilder{
		portfolioSvc: portfolioSvc,
		newsSvc:      newsSvc,
		insights:     insights,
	}
}

func (s *ReportBuilder) BuildDailyReport(
	ctx context.Context,
	portfolioName string,
	newsLimit int,
) (DailyReportResult, error) {
	portfolioAnalysis, err := s.portfolioSvc.GetAnalysis(ctx, portfolioName)
	if err != nil {
		return DailyReportResult{}, fmt.Errorf("get portfolio analysis: %w", err)
	}

	portfolioRisk, err := s.portfolioSvc.GetRisk(ctx, portfolioName)
	if err != nil {
		return DailyReportResult{}, fmt.Errorf("get portfolio risk: %w", err)
	}

	topPositions := s.insights.TopHoldings(portfolioAnalysis, 3)
	topHoldings := make([]TopHoldingReport, 0, len(topPositions))
	for _, pos := range topPositions {
		topHoldings = append(topHoldings, TopHoldingReport{
			Symbol:          pos.Symbol,
			Weight:          pos.Weight,
			MarketValueBase: pos.MarketValueBase,
			QuotedPrice:     pos.QuotedPrice,
			PriceCurrency:   pos.PriceCurrency,
			ChangePercent:   pos.QuotedChangePct,
		})
	}

	out := DailyReportResult{
		Report: DailyReport{
			PortfolioName: portfolioAnalysis.PortfolioName,
			BaseCurrency:  portfolioAnalysis.BaseCurrency,
			TotalValue:    portfolioAnalysis.TotalValue,
			TopHoldings:   topHoldings,
			Risk: RiskReport{
				LargestPosition:   portfolioRisk.LargestPosition,
				LargestWeight:     portfolioRisk.LargestWeight,
				Top3Concentration: portfolioRisk.Top3Concentration,
				Observations:      append([]string(nil), portfolioRisk.Observations...),
			},
			Attention: s.insights.AttentionSignals(portfolioAnalysis, portfolioRisk),
			News:      make([]TickerNewsReport, 0, len(topHoldings)),
		},
		Analysis: portfolioAnalysis,
		Risk:     portfolioRisk,
	}

	var wg sync.WaitGroup
	news := make([]TickerNewsReport, len(out.Report.TopHoldings))
	errCh := make(chan error, len(out.Report.TopHoldings))

	for i, holding := range out.Report.TopHoldings {
		wg.Add(1)

		go func(i int, symbol string) {
			defer wg.Done()

			n, err := s.newsSvc.GetNews(ctx, symbol, newsLimit)
			if err != nil {
				errCh <- err
				return
			}

			news[i] = n
		}(i, holding.Symbol)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return DailyReportResult{}, err
	}

	out.Report.News = news
	return out, nil
}
