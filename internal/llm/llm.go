package llm

import "context"

type LLMClient interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Ping(ctx context.Context) error
}

type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
}

type CompletionResponse struct {
	Text string
}
