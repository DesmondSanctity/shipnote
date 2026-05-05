package ai

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

//go:embed prompts/*.tmpl
var promptFS embed.FS

// renderPrompt loads the audience template and substitutes the
// canonical-JSON encoding of in.Facts. It returns the rendered string
// plus its sha256 (used as PromptHash in SummarizeOutput).
func renderPrompt(in SummarizeInput) (string, string, error) {
	name := fmt.Sprintf("prompts/%s.tmpl", in.Audience)
	body, err := promptFS.ReadFile(name)
	if err != nil {
		return "", "", fmt.Errorf("ai: load prompt %s: %w", name, err)
	}
	t, err := template.New(string(in.Audience)).Parse(string(body))
	if err != nil {
		return "", "", fmt.Errorf("ai: parse prompt %s: %w", name, err)
	}
	facts, err := canonicalJSON(in.Facts)
	if err != nil {
		return "", "", fmt.Errorf("ai: canonical facts: %w", err)
	}
	var sb strings.Builder
	if err := t.Execute(&sb, struct{ Facts string }{facts}); err != nil {
		return "", "", fmt.Errorf("ai: execute prompt: %w", err)
	}
	rendered := sb.String()
	sum := sha256.Sum256([]byte(rendered))
	return rendered, hex.EncodeToString(sum[:]), nil
}

// canonicalJSON marshals m with sorted map keys at every level so the
// resulting bytes are stable across runs and can be hashed for caching.
func canonicalJSON(m map[string]any) (string, error) {
	if m == nil {
		return "{}", nil
	}
	return marshalCanonical(m)
}

func marshalCanonical(v any) (string, error) {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var sb strings.Builder
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(',')
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return "", err
			}
			sb.Write(kb)
			sb.WriteByte(':')
			vs, err := marshalCanonical(x[k])
			if err != nil {
				return "", err
			}
			sb.WriteString(vs)
		}
		sb.WriteByte('}')
		return sb.String(), nil
	case []any:
		var sb strings.Builder
		sb.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				sb.WriteByte(',')
			}
			es, err := marshalCanonical(e)
			if err != nil {
				return "", err
			}
			sb.WriteString(es)
		}
		sb.WriteByte(']')
		return sb.String(), nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
