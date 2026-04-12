package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/llm"
	"github.com/squeakycheese75/tick/internal/report"
)

type LLMClient interface {
	Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error)
}

type ReportService interface {
	BuildDailyReport(ctx context.Context, portfolioName string, newsLimit int) (report.DailyReport, error)
}

type GetDailyReportUseCase struct {
	reportSvc ReportService
	llmClient LLMClient
}

func NewGetDailyReportUseCase(reportSvc ReportService, llmClient LLMClient) *GetDailyReportUseCase {
	return &GetDailyReportUseCase{
		reportSvc: reportSvc,
		llmClient: llmClient,
	}
}

func (uc *GetDailyReportUseCase) Execute(
	ctx context.Context,
	in GetDailyReportInput,
) (GetDailyReportOutput, error) {

	if in.PortfolioName == "" {
		in.PortfolioName = "main"
	}

	if in.NewsLimit <= 0 {
		in.NewsLimit = 2
	}

	dailyReport, err := uc.reportSvc.BuildDailyReport(
		ctx,
		in.PortfolioName,
		in.NewsLimit,
	)
	if err != nil {
		return GetDailyReportOutput{}, err
	}

	if !in.WithAI {
		return GetDailyReportOutput{
			DailyReport: dailyReport,
		}, nil
	}

	if uc.llmClient == nil {
		return GetDailyReportOutput{}, fmt.Errorf("ai not configured")
	}

	systemPrompt := buildDailyBriefAISystemPrompt()
	userPrompt := buildDailyBriefAIUserPrompt(dailyReport)

	resp, err := uc.llmClient.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	})
	if err != nil {
		return GetDailyReportOutput{}, fmt.Errorf("generate ai summary: %w", err)
	}

	return GetDailyReportOutput{
		DailyReport: dailyReport,
		AISummary:   strings.TrimSpace(resp.Text),
	}, nil
}

func buildDailyBriefAISystemPrompt() string {
	return strings.TrimSpace(`
You are a cautious portfolio analyst.

Your job is to summarize the supplied portfolio facts clearly and briefly.
Only use the information provided.
Do not invent any figures, positions, news, or market context.
Do not provide buy or sell recommendations.
Focus on concentration, notable daily moves, and what deserves attention today.
Return 3 to 5 concise bullet points.
`)
}

func buildDailyBriefAIUserPrompt(brief report.DailyReport) string {
	var b strings.Builder

	b.WriteString("Portfolio daily brief\n\n")

	b.WriteString(fmt.Sprintf("Portfolio: %s\n", brief.PortfolioName))
	b.WriteString(fmt.Sprintf("Base currency: %s\n", brief.BaseCurrency))
	b.WriteString(fmt.Sprintf("Total value: %.2f %s\n\n", brief.TotalValue, brief.BaseCurrency))

	b.WriteString("Top holdings:\n")
	if len(brief.TopHoldings) == 0 {
		b.WriteString("- No positions\n")
	} else {
		for _, h := range brief.TopHoldings {
			b.WriteString(fmt.Sprintf(
				"- %s: weight %.2f%%, value %.2f %s, quoted price %.2f %s, daily move %+.2f%%\n",
				h.Symbol,
				h.Weight*100,
				h.MarketValueBase,
				brief.BaseCurrency,
				h.QuotedPrice,
				h.PriceCurrency,
				h.ChangePercent,
			))
		}
	}
	b.WriteString("\n")

	b.WriteString("Risk:\n")
	if brief.Risk.LargestPosition == "" {
		b.WriteString("- No risk data available\n")
	} else {
		b.WriteString(fmt.Sprintf("- Largest position: %s (%.2f%%)\n", brief.Risk.LargestPosition, brief.Risk.LargestWeight*100))
		b.WriteString(fmt.Sprintf("- Top 3 concentration: %.2f%%\n", brief.Risk.Top3Concentration*100))
		for _, obs := range brief.Risk.Observations {
			b.WriteString(fmt.Sprintf("- %s\n", obs))
		}
	}
	b.WriteString("\n")

	b.WriteString("News:\n")
	if len(brief.News) == 0 {
		b.WriteString("- No news available\n")
	} else {
		for _, group := range brief.News {
			if len(group.Headlines) == 0 {
				b.WriteString(fmt.Sprintf("- %s: no recent headlines\n", group.Ticker))
				continue
			}
			b.WriteString(fmt.Sprintf("- %s:\n", group.Ticker))
			for _, headline := range group.Headlines {
				b.WriteString(fmt.Sprintf("  - %s\n", headline.Title))
			}
		}
	}
	b.WriteString("\n")

	b.WriteString("Attention:\n")
	if len(brief.Attention) == 0 {
		b.WriteString("- None\n")
	} else {
		for _, item := range brief.Attention {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}
	b.WriteString("\n")

	b.WriteString("Task:\n")
	b.WriteString("Summarize what matters today in 3 to 5 concise bullet points.\n")

	return b.String()
}
