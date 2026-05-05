package ai

import (
	"context"
	"fmt"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// Options bundles knobs for a single Enrich pass. Audiences defaults
// to AllAudiences() when empty. Cache may be nil.
type Options struct {
	Provider  Summarizer
	Cache     *Cache
	Audiences []Audience
	MaxTokens int
	// Logger receives non-fatal messages (provider unavailable, single
	// summary failed). nil disables logging.
	Logger func(msg string)
}

// Enrich populates rel.Summary.AISummaries and per-Change AISummaries
// using opts.Provider for every audience in opts.Audiences. A nil or
// unavailable provider is a no-op (deterministic baseline preserved).
//
// Any single-summary failure is logged and skipped; the rest of the
// release is still enriched. AI is additive — never fatal.
func Enrich(ctx context.Context, rel *model.Release, opts Options) error {
	if rel == nil || opts.Provider == nil {
		return nil
	}
	audiences := opts.Audiences
	if len(audiences) == 0 {
		audiences = AllAudiences()
	}
	ok, err := opts.Provider.Available(ctx)
	if err != nil || !ok {
		log(opts.Logger, fmt.Sprintf("ai: provider %s unavailable, skipping", opts.Provider.Name()))
		return nil //nolint:nilerr // unavailable is non-fatal by design
	}

	// Release-level summary.
	releaseFacts := releaseFacts(rel)
	rel.Summary.AISummaries = summariseAll(ctx, opts, audiences, "release", releaseFacts)

	// Per-change summaries.
	for i := range rel.GlobalChanges {
		ch := &rel.GlobalChanges[i]
		ch.AISummaries = summariseAll(ctx, opts, audiences,
			"change:"+ch.ID, changeFacts(ch))
	}
	return nil
}

func summariseAll(ctx context.Context, opts Options, audiences []Audience, scope string, facts map[string]any) model.AISummaries {
	out := model.AISummaries{}
	for _, a := range audiences {
		s, ok := summarise(ctx, opts, SummarizeInput{
			Audience:  a,
			Scope:     scope,
			Facts:     facts,
			MaxTokens: opts.MaxTokens,
		})
		if !ok {
			continue
		}
		entry := &model.AISummary{
			Text:          s.Text,
			Model:         s.Model,
			PromptVersion: PromptVersion,
		}
		switch a {
		case AudienceDeveloper:
			out.Developer = entry
		case AudienceCustomer:
			out.Customer = entry
		case AudienceMigration:
			out.Migration = entry
		}
	}
	return out
}

func summarise(ctx context.Context, opts Options, in SummarizeInput) (SummarizeOutput, bool) {
	_, hash, err := renderPrompt(in)
	if err != nil {
		log(opts.Logger, fmt.Sprintf("ai: render prompt %s: %v", in.Audience, err))
		return SummarizeOutput{}, false
	}
	if hit, ok := opts.Cache.Lookup(opts.Provider.Name(), modelOf(opts.Provider), in, hash); ok {
		return hit, true
	}
	out, err := opts.Provider.Summarize(ctx, in)
	if err != nil {
		log(opts.Logger, fmt.Sprintf("ai: summarise %s scope=%s: %v", in.Audience, in.Scope, err))
		return SummarizeOutput{}, false
	}
	if err := opts.Cache.Save(opts.Provider.Name(), out.Model, in, out); err != nil {
		log(opts.Logger, fmt.Sprintf("ai: cache save: %v", err))
	}
	return out, true
}

// modelOf reaches inside an OpenAICompatible to read the configured
// model so the cache key is stable before the first network call.
// Other Summarizer implementations would expose the model differently;
// for v1 every provider is OpenAICompatible-shaped.
func modelOf(s Summarizer) string {
	if oc, ok := s.(*OpenAICompatible); ok {
		return oc.cfg.Model
	}
	return ""
}

func log(fn func(string), msg string) {
	if fn != nil {
		fn(msg)
	}
}

// releaseFacts pulls the deterministic, non-AI fields the templates
// need. Keep this small: prompts that see too much context wander.
func releaseFacts(rel *model.Release) map[string]any {
	changes := make([]map[string]any, 0, len(rel.GlobalChanges))
	for _, c := range rel.GlobalChanges {
		changes = append(changes, map[string]any{
			"type":     string(c.Type),
			"breaking": c.Breaking,
			"title":    c.Title,
		})
	}
	return map[string]any{
		"repo":            rel.Repo.Owner + "/" + rel.Repo.Name,
		"tag":             rel.Release.Tag,
		"state":           string(rel.Release.State),
		"stats":           map[string]any{"changes": rel.Stats.Changes, "breaking": rel.Stats.Breaking, "fixes": rel.Stats.Fixes, "features": rel.Stats.Features},
		"changes":         changes,
		"breakingChanges": rel.BreakingChanges,
	}
}

// changeFacts is the per-change payload. Author and PR URL are
// excluded because the renderer adds those literally; including them
// here just tempts the model to re-emit them as filler text.
func changeFacts(c *model.Change) map[string]any {
	out := map[string]any{
		"type":     string(c.Type),
		"breaking": c.Breaking,
		"title":    c.Title,
	}
	if c.Description != "" {
		out["description"] = c.Description
	}
	if c.Migration != nil && c.Migration.Summary != "" {
		out["migration"] = c.Migration.Summary
	}
	if len(c.Packages) > 0 {
		out["packages"] = c.Packages
	}
	return out
}
