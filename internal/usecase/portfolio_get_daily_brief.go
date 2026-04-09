package usecase

import (
	"context"
	"fmt"

	"github.com/squeakycheese75/tick/internal/domain"
)

type GetDailyBriefUseCase struct {
	portfolioService  PortfolioSvc
	portfolioInsights PortfolioInsights
	newsProvider      NewsProvider
}

func NewGetDailyBriefUseCase(
	portfolioService PortfolioSvc,
	portfolioInsights PortfolioInsights,
	newsProvider NewsProvider,
) *GetDailyBriefUseCase {
	return &GetDailyBriefUseCase{
		portfolioService:  portfolioService,
		portfolioInsights: portfolioInsights,
		newsProvider:      newsProvider,
	}
}

func (uc *GetDailyBriefUseCase) Execute(
	ctx context.Context,
	in GetDailyBriefInput,
) (GetDailyBriefOutput, error) {
	if in.PortfolioName == "" {
		in.PortfolioName = "main"
	}

	if in.NewsLimit <= 0 {
		in.NewsLimit = 2
	}

	portfolioAnalysis, err := uc.portfolioService.GetAnalysis(ctx, in.PortfolioName)
	if err != nil {
		return GetDailyBriefOutput{}, fmt.Errorf("get portfolio analysis: %w", err)
	}

	portfolioRisk, err := uc.portfolioService.GetRisk(ctx, in.PortfolioName)
	if err != nil {
		return GetDailyBriefOutput{}, fmt.Errorf("get portfolio risk: %w", err)
	}

	topPositions := uc.portfolioInsights.TopHoldings(portfolioAnalysis, 3)
	topHoldings := make([]DailyHolding, 0, len(topPositions))
	for _, pos := range topPositions {
		topHoldings = append(topHoldings, DailyHolding{
			Ticker:          pos.Ticker,
			Weight:          pos.Weight,
			MarketValueBase: pos.MarketValueBase,
			QuotedPrice:     pos.QuotedPrice,
			PriceCurrency:   pos.PriceCurrency,
			ChangePercent:   pos.QuotedChangePct,
		})
	}

	out := GetDailyBriefOutput{
		PortfolioName: portfolioAnalysis.PortfolioName,
		BaseCurrency:  portfolioAnalysis.BaseCurrency,
		TotalValue:    portfolioAnalysis.TotalValue,
		TopHoldings:   topHoldings,
		Risk: DailyRisk{
			LargestPosition:   portfolioRisk.LargestPosition,
			LargestWeight:     portfolioRisk.LargestWeight,
			Top3Concentration: portfolioRisk.Top3Concentration,
			Observations:      append([]string(nil), portfolioRisk.Observations...),
		},
		Attention: uc.portfolioInsights.AttentionSignals(portfolioAnalysis, portfolioRisk),
		News:      make([]DailyNews, 0),
	}

	for _, holding := range out.TopHoldings {
		headlines, err := uc.newsProvider.GetNews(ctx, holding.Ticker, in.NewsLimit)
		if err != nil {
			return GetDailyBriefOutput{}, fmt.Errorf("get news for %s: %w", holding.Ticker, err)
		}

		out.News = append(out.News, DailyNews{
			Ticker:    holding.Ticker,
			Headlines: mapNewsHeadlines(headlines),
		})
	}

	return out, nil
}

func mapNewsHeadlines(in []domain.NewsHeadline) []NewsHeadline {
	out := make([]NewsHeadline, 0, len(in))
	for _, h := range in {
		out = append(out, NewsHeadline{
			Title: h.Title,
			URL:   h.URL,
		})
	}
	return out
}
