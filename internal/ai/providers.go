package ai

import (
	"net/http"
	"os"
	"strings"
	"time"
)

// NewOllama returns a Summarizer for a locally running Ollama daemon.
// Ollama exposes an OpenAI-compatible endpoint at /v1; no API key is
// required. Override the address with the OLLAMA_HOST environment
// variable (e.g. "http://192.168.1.10:11434"); default is localhost.
func NewOllama(model string) *OpenAICompatible {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	return NewOpenAICompatible(OpenAICompatibleConfig{
		Name:    "ollama",
		BaseURL: strings.TrimRight(host, "/") + "/v1",
		Model:   model,
		HTTP:    &http.Client{Timeout: 60 * time.Second}, // local models are slower
	})
}

// NewOpenAI returns a Summarizer for api.openai.com. The token is read
// from the supplied argument; pass os.Getenv("OPENAI_API_KEY") at the
// call site. Empty token returns a client that will fail Available.
func NewOpenAI(token, model string) *OpenAICompatible {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return NewOpenAICompatible(OpenAICompatibleConfig{
		Name:    "openai",
		BaseURL: "https://api.openai.com/v1",
		APIKey:  token,
		Model:   model,
	})
}

// NewGroq returns a Summarizer for api.groq.com. Groq hosts open
// models behind an OpenAI-compatible API and has a free tier suitable
// for OSS-project release notes.
func NewGroq(token, model string) *OpenAICompatible {
	if model == "" {
		model = "llama-3.1-8b-instant"
	}
	return NewOpenAICompatible(OpenAICompatibleConfig{
		Name:    "groq",
		BaseURL: "https://api.groq.com/openai/v1",
		APIKey:  token,
		Model:   model,
	})
}
