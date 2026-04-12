package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (c *OllamaClient) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	prompt := buildOllamaPrompt(req)

	payload := ollamaGenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/generate",
		bytes.NewReader(body),
	)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("build ollama request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("perform ollama request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return CompletionResponse{}, fmt.Errorf("ollama returned status %s", resp.Status)
	}

	var data ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return CompletionResponse{}, fmt.Errorf("decode ollama response: %w", err)
	}

	return CompletionResponse{
		Text: strings.TrimSpace(data.Response),
	}, nil
}

func (c *OllamaClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/api/tags",
		nil,
	)
	if err != nil {
		return fmt.Errorf("build ollama ping request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("connect to ollama: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama not healthy: %s", resp.Status)
	}

	var data struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("decode ollama ping response: %w", err)
	}

	want := normalizeModel(c.model)

	for _, m := range data.Models {
		if normalizeModel(m.Name) == want {
			return nil
		}
	}

	return fmt.Errorf("model %s not found in ollama", c.model)
}

func buildOllamaPrompt(req CompletionRequest) string {
	var b strings.Builder

	if strings.TrimSpace(req.SystemPrompt) != "" {
		b.WriteString("System:\n")
		b.WriteString(strings.TrimSpace(req.SystemPrompt))
		b.WriteString("\n\n")
	}

	if strings.TrimSpace(req.UserPrompt) != "" {
		b.WriteString("User:\n")
		b.WriteString(strings.TrimSpace(req.UserPrompt))
	}

	return b.String()
}

func normalizeModel(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if !strings.Contains(name, ":") {
		return name + ":latest"
	}
	return name
}
