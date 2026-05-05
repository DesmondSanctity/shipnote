package analyze

import (
	"regexp"
	"strings"
)

// BreakingDecision flags whether a change is breaking and why.
type BreakingDecision struct {
	Breaking   bool
	Confidence float64
	Source     string // "label", "cc-bang", "footer", "default"
	Notice     string // text after "BREAKING CHANGE:" footer when present
}

// DetectBreaking inspects the same Input used by Categorize and decides
// if the change is API-breaking. Any positive signal wins; negative
// signals don't override a positive one.
//
//  1. Label matching "breaking", "breaking-change", "semver:major"
//  2. Conventional-commit "!" marker in the title or commit message
//  3. "BREAKING CHANGE:" / "BREAKING-CHANGE:" footer in PR body
//  4. Default: not breaking
func DetectBreaking(in Input) BreakingDecision {
	footer, hasFooter := breakingFromFooter(in.PRBody)
	if d, ok := breakingFromLabels(in.Labels); ok {
		if hasFooter {
			d.Notice = footer.Notice
		}
		return d
	}
	if d, ok := breakingFromCC(firstNonEmpty(in.PRTitle, in.CommitMsg)); ok {
		if hasFooter {
			d.Notice = footer.Notice
		}
		return d
	}
	if hasFooter {
		return footer
	}
	return BreakingDecision{Breaking: false, Confidence: 0.9, Source: "default"}
}

var breakingLabelSet = map[string]struct{}{
	"breaking":        {},
	"breaking-change": {},
	"breaking change": {},
	"semver:major":    {},
	"semver-major":    {},
	"major":           {},
}

func breakingFromLabels(labels []string) (BreakingDecision, bool) {
	for _, raw := range labels {
		key := canonLabel(raw)
		if _, ok := breakingLabelSet[key]; ok {
			return BreakingDecision{Breaking: true, Confidence: 0.95, Source: "label"}, true
		}
	}
	return BreakingDecision{}, false
}

// ccBangRe matches a CC type with the "!" breaking marker.
var ccBangRe = regexp.MustCompile(`^[a-zA-Z]+(?:\([^)]+\))?!:\s+`)

func breakingFromCC(title string) (BreakingDecision, bool) {
	if ccBangRe.MatchString(strings.TrimSpace(title)) {
		return BreakingDecision{Breaking: true, Confidence: 0.9, Source: "cc-bang"}, true
	}
	return BreakingDecision{}, false
}

// breakingFooterRe matches the trailing footer block. Per the
// Conventional Commits spec the footer must start at the beginning of a
// line, optionally with "BREAKING-CHANGE" or "BREAKING CHANGE", and
// continues until a blank line or end of input.
var breakingFooterRe = regexp.MustCompile(`(?im)^breaking[\s-]change\s*:\s*(.+)$`)

func breakingFromFooter(body string) (BreakingDecision, bool) {
	m := breakingFooterRe.FindStringSubmatch(body)
	if m == nil {
		return BreakingDecision{}, false
	}
	notice := strings.TrimSpace(m[1])
	return BreakingDecision{
		Breaking:   true,
		Confidence: 0.95,
		Source:     "footer",
		Notice:     notice,
	}, true
}
