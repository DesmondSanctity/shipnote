package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatibleConfig configures any provider that speaks the
// OpenAI Chat Completions wire format. This single client covers
// OpenAI, Groq, OpenRouter, Together, vLLM, LM Studio and Ollama's
// /v1 compatibility endpoint via BaseURL switching.
//
// BaseURL must include the trailing /v1 (e.g. https://api.openai.com/v1).
type OpenAICompatibleConfig struct {
	Name    string // human-readable provider name, e.g. "openai", "groq"
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
	Timeout time.Duration
}

// OpenAICompatible is a Summarizer for any OpenAI-API-shaped backend.
type OpenAICompatible struct {
	cfg OpenAICompatibleConfig
}

// NewOpenAICompatible returns a configured client. APIKey may be empty
// for providers that don't require auth (some local servers).
func NewOpenAICompatible(cfg OpenAICompatibleConfig) *OpenAICompatible {
	if cfg.HTTP == nil {
		cfg.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &OpenAICompatible{cfg: cfg}
}

// Name returns a stable identifier used in cache keys.
func (o *OpenAICompatible) Name() string {
	if o.cfg.Name != "" {
		return o.cfg.Name
	}
	return "openai-compatible"
}

// Available probes the /models endpoint to confirm the server is up
// and the API key (if any) is accepted.
func (o *OpenAICompatible) Available(ctx context.Context) (bool, error) {
	if o.cfg.BaseURL == "" {
		return false, ErrUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.cfg.BaseURL+"/models", nil)
	if err != nil {
		return false, err
	}
	if o.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)
	}
	resp, err := o.cfg.HTTP.Do(req)
	if err != nil {
		return false, ErrUnavailable
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

type openaiChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Summarize renders the audience template, POSTs to /chat/completions,
// and returns the model's first-choice text.
func (o *OpenAICompatible) Summarize(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
	prompt, hash, err := renderPrompt(in)
	if err != nil {
		return SummarizeOutput{}, err
	}
	body, err := json.Marshal(openaiChatRequest{
		Model:       o.cfg.Model,
		Temperature: 0, // deterministic path; cache is the second guard
		MaxTokens:   in.MaxTokens,
		Messages: []openaiMessage{
			{Role: "system", Content: "You produce release notes. Follow the user's rules exactly."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return SummarizeOutput{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return SummarizeOutput{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if o.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)
	}
	resp, err := o.cfg.HTTP.Do(req)
	if err != nil {
		return SummarizeOutput{}, fmt.Errorf("ai: %s: %w", o.Name(), err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return SummarizeOutput{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SummarizeOutput{}, fmt.Errorf("ai: %s http %d: %s", o.Name(), resp.StatusCode, truncate(string(raw), 200))
	}
	var cr openaiChatResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return SummarizeOutput{}, fmt.Errorf("ai: %s decode: %w", o.Name(), err)
	}
	if cr.Error != nil {
		return SummarizeOutput{}, fmt.Errorf("ai: %s: %s", o.Name(), cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return SummarizeOutput{}, fmt.Errorf("ai: %s: empty choices", o.Name())
	}
	model := cr.Model
	if model == "" {
		model = o.cfg.Model
	}
	return SummarizeOutput{
		Text:       strings.TrimSpace(cr.Choices[0].Message.Content),
		Model:      model,
		PromptHash: hash,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
