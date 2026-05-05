package ai

import (
	"context"
	"errors"
)

// PromptVersion pins the active prompt template set. Bump when a
// change to a prompt alters output meaning (rules, structure, format).
// Cosmetic edits (typos, whitespace) do not bump.
//
// The cache key includes this string so a bump invalidates stale
// summaries automatically on the next run with the new binary.
const PromptVersion = "v1"

// Audience is the render-time tone of an AI summary. Each audience
// has its own prompt template; the model never picks tone freely.
type Audience string

// Audience values are stable strings used as map keys, prompt
// filenames, and cache-key components; do not rename without bumping
// PromptVersion.
const (
	AudienceDeveloper Audience = "developer"
	AudienceCustomer  Audience = "customer"
	AudienceMigration Audience = "migration"
)

// AllAudiences returns the audience values in deterministic order.
func AllAudiences() []Audience {
	return []Audience{AudienceDeveloper, AudienceCustomer, AudienceMigration}
}

// SummarizeInput is the structured fact payload sent to a provider.
// Facts is canonical (sorted keys, stable values) so the cache key is
// reproducible across runs.
type SummarizeInput struct {
	Audience  Audience
	Scope     string         // "release" | "package:<id>" | "change:<id>"
	Facts     map[string]any // structured facts ONLY, never raw diffs
	MaxTokens int            // soft cap; provider may clamp
}

// SummarizeOutput is the rendered prose plus traceability fields.
type SummarizeOutput struct {
	Text       string
	Model      string
	PromptHash string // sha256 of the rendered prompt
}

// Summarizer is the provider interface. Implementations live in this
// package and must respect determinism: same canonical input → same
// output (the cache layer enforces this for free).
type Summarizer interface {
	Name() string
	Available(ctx context.Context) (bool, error)
	Summarize(ctx context.Context, in SummarizeInput) (SummarizeOutput, error)
}

// ErrUnavailable signals a provider is reachable conceptually but not
// usable in this environment (no token, server down, model missing).
// Callers should treat it as "skip this provider, try the next" — not
// as a hard failure.
var ErrUnavailable = errors.New("ai: provider unavailable")
