package analyze

import (
	"sort"
	"strings"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// IdentityRule maps one or more aliases (logins, emails, names) to a
// single canonical contributor identity. Matched case-insensitively.
//
// Typical config-driven entry:
//
//	[[authors]]
//	canonical = { login = "maya", name = "Maya R.", url = "..." }
//	aliases   = ["maya@old.example", "Maya Rivera"]
type IdentityRule struct {
	Canonical model.Contributor
	Aliases   []string
}

// IdentityMerger applies a list of IdentityRule to incoming
// Contributor records and aggregates Changes counts under the
// canonical identity. The zero value (nil rules) acts as identity:
// each unique Login is kept as-is.
type IdentityMerger struct {
	rules []IdentityRule
	// alias -> rule index, lowercased keys
	index map[string]int
}

// NewIdentityMerger compiles rules into an alias index. Each alias and
// the canonical Login itself become lookup keys.
func NewIdentityMerger(rules []IdentityRule) *IdentityMerger {
	m := &IdentityMerger{
		rules: append([]IdentityRule(nil), rules...),
		index: make(map[string]int, len(rules)*2),
	}
	for i, r := range rules {
		if r.Canonical.Login != "" {
			m.index[strings.ToLower(r.Canonical.Login)] = i
		}
		for _, a := range r.Aliases {
			a = strings.ToLower(strings.TrimSpace(a))
			if a != "" {
				m.index[a] = i
			}
		}
	}
	return m
}

// Resolve returns the canonical Contributor for the given input. If no
// rule matches, the input is returned unchanged with its Changes count
// preserved.
func (m *IdentityMerger) Resolve(c model.Contributor) model.Contributor {
	if m == nil {
		return c
	}
	for _, key := range []string{c.Login, c.Name} {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			continue
		}
		if i, ok := m.index[key]; ok {
			out := m.rules[i].Canonical
			out.Changes = c.Changes
			return out
		}
	}
	return c
}

// MergeContributors reduces a slice of Contributor entries into a
// deduplicated, sorted list. Records sharing a canonical Login (after
// applying the merger) have their Changes counts summed. Output order
// is descending by Changes, then ascending by Login for stability.
func MergeContributors(merger *IdentityMerger, in []model.Contributor) []model.Contributor {
	if len(in) == 0 {
		return nil
	}
	bucket := make(map[string]model.Contributor, len(in))
	for _, c := range in {
		canon := c
		if merger != nil {
			canon = merger.Resolve(c)
		}
		key := strings.ToLower(canon.Login)
		if key == "" {
			key = strings.ToLower(canon.Name)
		}
		if existing, ok := bucket[key]; ok {
			existing.Changes += c.Changes
			bucket[key] = existing
		} else {
			canon.Changes = c.Changes
			bucket[key] = canon
		}
	}
	out := make([]model.Contributor, 0, len(bucket))
	for _, c := range bucket {
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Changes != out[j].Changes {
			return out[i].Changes > out[j].Changes
		}
		return strings.ToLower(out[i].Login) < strings.ToLower(out[j].Login)
	})
	return out
}
