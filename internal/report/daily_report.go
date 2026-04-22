package report

import (
	"context"
	"fmt"
	"sync"

	"github.com/squeakycheese75/tick/internal/domain"
)

type (
	PortfolioService interface {
		GetAnalysis(ctx context.Context, portfolioName string) (domain.PortfolioAnalysis, error)
		GetRisk(ctx context.Context, portfolioName string) (domain.PortfolioRisk, error)
	}
	NewsService interface {
		GetNews(ctx context.Context, ticker string, newsLimit int) (domain.TickerNewsReport, error)
	}
	PortfolioInsights interface {
		TopHoldings(portfolioAnalysis domain.PortfolioAnalysis, limit int) []domain.AnalyzedPosition
		AttentionSignals(portfolioAnalysis domain.PortfolioAnalysis, portfolioRisk domain.PortfolioRisk) []string
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
) (domain.DailyReportResult, error) {
	portfolioAnalysis, err := s.portfolioSvc.GetAnalysis(ctx, portfolioName)
	if err != nil {
		return domain.DailyReportResult{}, fmt.Errorf("get portfolio analysis: %w", err)
	}

	portfolioRisk, err := s.portfolioSvc.GetRisk(ctx, portfolioName)
	if err != nil {
		return domain.DailyReportResult{}, fmt.Errorf("get portfolio risk: %w", err)
	}

	topPositions := s.insights.TopHoldings(portfolioAnalysis, 3)
	topHoldings := make([]domain.TopHoldingReport, 0, len(topPositions))
	for _, pos := range topPositions {
		topHoldings = append(topHoldings, domain.TopHoldingReport{
			Symbol:          pos.Symbol,
			Weight:          pos.Weight,
			MarketValueBase: pos.MarketValueBase,
			QuotedPrice:     pos.QuotedPrice,
			PriceCurrency:   pos.PriceCurrency,
			ChangePercent:   pos.QuotedChangePct,
		})
	}

	out := domain.DailyReportResult{
		Report: domain.DailyReport{
			PortfolioName: portfolioAnalysis.PortfolioName,
			BaseCurrency:  portfolioAnalysis.BaseCurrency,
			TotalValue:    portfolioAnalysis.TotalValue,
			TopHoldings:   topHoldings,
			Risk: domain.RiskReport{
				LargestPosition:   portfolioRisk.LargestPosition,
				LargestWeight:     portfolioRisk.LargestWeight,
				Top3Concentration: portfolioRisk.Top3Concentration,
				Observations:      append([]string(nil), portfolioRisk.Observations...),
			},
			Attention: s.insights.AttentionSignals(portfolioAnalysis, portfolioRisk),
			News:      make([]domain.TickerNewsReport, 0, len(topHoldings)),
		},
		Analysis: portfolioAnalysis,
		Risk:     portfolioRisk,
	}

	var wg sync.WaitGroup
	news := make([]domain.TickerNewsReport, len(out.Report.TopHoldings))
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
		return domain.DailyReportResult{}, err
	}

	out.Report.News = news
	return out, nil
}

func EnrichDailyReportWithSnapshot(
	dailyReport domain.DailyReport,
	previousSnapshot domain.PortfolioSnapshot,
	previousPositions []domain.PortfolioSnapshotPosition,
) domain.DailyReport {
	delta := dailyReport.TotalValue - previousSnapshot.TotalValue

	var pct float64
	if previousSnapshot.TotalValue != 0 {
		pct = delta / previousSnapshot.TotalValue
	}

	dailyReport.ChangeSinceLastSnapshot = &domain.ValueChangeReport{
		Absolute: delta,
		Percent:  pct,
	}

	previousBySymbol := make(map[string]domain.PortfolioSnapshotPosition, len(previousPositions))
	for _, p := range previousPositions {
		previousBySymbol[p.Symbol] = p
	}

	for i := range dailyReport.TopHoldings {
		current := dailyReport.TopHoldings[i]
		previous, ok := previousBySymbol[current.Symbol]
		if !ok {
			continue
		}

		valueDelta := current.MarketValueBase - previous.MarketValueBase

		var valuePct float64
		if previous.MarketValueBase != 0 {
			valuePct = valueDelta / previous.MarketValueBase
		}

		dailyReport.TopHoldings[i].SinceLastSnapshot = &domain.HoldingSnapshotChangeReport{
			ValueAbsolute: valueDelta,
			ValuePercent:  valuePct,
		}
	}

	return dailyReport
}
