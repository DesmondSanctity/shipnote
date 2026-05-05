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

// AnthropicConfig configures a native Anthropic Messages API client.
// Anthropic does not speak the OpenAI Chat Completions wire format,
// so it gets its own thin adapter rather than reusing OpenAICompatible.
//
// BaseURL must include /v1 (default https://api.anthropic.com/v1).
type AnthropicConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Version string // anthropic-version header; defaults to "2023-06-01"
	HTTP    *http.Client
}

// Anthropic is a Summarizer for api.anthropic.com.
type Anthropic struct {
	cfg AnthropicConfig
}

// NewAnthropic returns a configured client. APIKey is required.
func NewAnthropic(token, model string) *Anthropic {
	if model == "" {
		model = "claude-3-5-haiku-latest"
	}
	return &Anthropic{cfg: AnthropicConfig{
		BaseURL: "https://api.anthropic.com/v1",
		APIKey:  token,
		Model:   model,
		Version: "2023-06-01",
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}}
}

// Name returns the provider id used in cache keys.
func (a *Anthropic) Name() string { return "anthropic" }

// SetBaseURLForTest overrides the base URL. Test-only helper kept on
// the public type so external _test packages can point at httptest
// servers without exposing the full config.
func (a *Anthropic) SetBaseURLForTest(u string) { a.cfg.BaseURL = u }

// Available probes /models. Anthropic returns 401 without a valid key,
// so this also validates auth before we spend tokens on a real call.
func (a *Anthropic) Available(ctx context.Context) (bool, error) {
	if a.cfg.APIKey == "" {
		return false, ErrUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.cfg.BaseURL+"/models", nil)
	if err != nil {
		return false, err
	}
	a.setHeaders(req)
	resp, err := a.cfg.HTTP.Do(req)
	if err != nil {
		return false, ErrUnavailable
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model string `json:"model"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Summarize renders the audience template, POSTs to /messages, and
// returns the first text block.
func (a *Anthropic) Summarize(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
	prompt, hash, err := renderPrompt(in)
	if err != nil {
		return SummarizeOutput{}, err
	}
	maxTokens := in.MaxTokens
	if maxTokens <= 0 {
		// Anthropic requires max_tokens; pick a value comfortably
		// above our prompt-mandated word limits.
		maxTokens = 512
	}
	body, err := json.Marshal(anthropicRequest{
		Model:     a.cfg.Model,
		MaxTokens: maxTokens,
		System:    "You produce release notes. Follow the user's rules exactly.",
		Messages:  []anthropicMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return SummarizeOutput{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return SummarizeOutput{}, err
	}
	a.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.cfg.HTTP.Do(req)
	if err != nil {
		return SummarizeOutput{}, fmt.Errorf("ai: anthropic: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return SummarizeOutput{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SummarizeOutput{}, fmt.Errorf("ai: anthropic http %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var ar anthropicResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return SummarizeOutput{}, fmt.Errorf("ai: anthropic decode: %w", err)
	}
	if ar.Error != nil {
		return SummarizeOutput{}, fmt.Errorf("ai: anthropic: %s", ar.Error.Message)
	}
	text := firstText(ar.Content)
	if text == "" {
		return SummarizeOutput{}, fmt.Errorf("ai: anthropic: empty content")
	}
	model := ar.Model
	if model == "" {
		model = a.cfg.Model
	}
	return SummarizeOutput{
		Text:       strings.TrimSpace(text),
		Model:      model,
		PromptHash: hash,
	}, nil
}

func (a *Anthropic) setHeaders(req *http.Request) {
	req.Header.Set("x-api-key", a.cfg.APIKey)
	v := a.cfg.Version
	if v == "" {
		v = "2023-06-01"
	}
	req.Header.Set("anthropic-version", v)
}

func firstText(blocks []struct {
	Type string `json:"type"`
	Text string `json:"text"`
},
) string {
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			return b.Text
		}
	}
	return ""
}
