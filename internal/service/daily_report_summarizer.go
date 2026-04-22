package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/squeakycheese75/tick/internal/domain"
	"github.com/squeakycheese75/tick/internal/llm"
)

type LLMClient interface {
	Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error)
}

type LLMDailyReportSummarizer struct {
	client LLMClient
}

func NewLLMDailyReportSummarizer(llmClient LLMClient) *LLMDailyReportSummarizer {
	return &LLMDailyReportSummarizer{
		client: llmClient,
	}
}

func (s *LLMDailyReportSummarizer) Summarize(
	ctx context.Context,
	brief domain.DailyReport,
) (string, error) {

	resp, err := s.client.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: buildDailyReportSystemPrompt(),
		UserPrompt:   buildDailyReportUserPrompt(brief),
	})
	if err != nil {
		return "", fmt.Errorf("generate ai summary: %w", err)
	}

	return strings.TrimSpace(resp.Text), nil
}

func (LLMDailyReportSummarizer) Enabled() bool {
	return true
}

type NoopSummarizer struct{}

func (NoopSummarizer) Summarize(ctx context.Context, r domain.DailyReport) (string, error) {
	return "", nil
}

func (NoopSummarizer) Enabled() bool {
	return false
}
