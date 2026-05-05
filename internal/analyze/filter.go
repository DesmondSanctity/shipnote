package analyze

import (
	"strings"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// Filter decides whether a Change should be included in the release.
// When excluded it returns the reason code recorded in
// Release.Excluded so consumers can audit what was dropped.
type Filter struct {
	// IgnoreBots drops changes whose only author is a bot account.
	IgnoreBots bool
	// IgnoreLogins is a set of author logins to drop (case-insensitive).
	IgnoreLogins map[string]struct{}
	// IgnoreLabels drops a change if any of its source PR labels match
	// (canonical form, see canonLabel).
	IgnoreLabels map[string]struct{}
	// DropChoreOnly drops Changes typed "chore" with no breaking flag.
	// Used by the "developer" audience preset; off by default so the
	// default render still mentions chore work in the audit trail.
	DropChoreOnly bool
}

// Decision is the output of Filter.Apply: either keep the change, or
// drop it with a stable reason.
type Decision struct {
	Keep   bool
	Reason string // "bot", "ignored-login", "ignored-label", "chore-only"
}

// Apply returns Keep=true to include a change, or Keep=false plus a
// stable reason code when the change should be excluded.
//
// labels are the raw PR labels (will be canonicalized internally).
// authorIsBot is the authoritative bot flag from the source adapter.
func (f Filter) Apply(c model.Change, labels []string, authorIsBot bool) Decision {
	if f.IgnoreBots && authorIsBot {
		return Decision{Reason: "bot"}
	}
	if len(f.IgnoreLogins) > 0 {
		if _, ok := f.IgnoreLogins[strings.ToLower(c.Author.Login)]; ok {
			return Decision{Reason: "ignored-login"}
		}
	}
	if len(f.IgnoreLabels) > 0 {
		for _, raw := range labels {
			if _, ok := f.IgnoreLabels[canonLabel(raw)]; ok {
				return Decision{Reason: "ignored-label"}
			}
		}
	}
	if f.DropChoreOnly && !c.Breaking && c.Type == model.ChangeTypeChore {
		return Decision{Reason: "chore-only"}
	}
	return Decision{Keep: true}
}

// LowerSet builds a case-insensitive lookup set from a slice of strings.
// Trims whitespace and skips empties.
func LowerSet(xs []string) map[string]struct{} {
	if len(xs) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(xs))
	for _, x := range xs {
		x = strings.ToLower(strings.TrimSpace(x))
		if x != "" {
			out[x] = struct{}{}
		}
	}
	return out
}
