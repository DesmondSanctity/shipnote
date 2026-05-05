package analyze

import (
	"regexp"
	"strings"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// CategoryDecision is the analyzer's choice for a single change, with a
// breadcrumb explaining how it was reached. Confidence is in [0, 1].
type CategoryDecision struct {
	Type       model.ChangeType
	Confidence float64
	Source     string // "label", "changelog-block", "cc-prefix", "heuristic", "default"
}

// Input is the minimal shape Categorize needs — a PR with optional
// labels and the headline commit message. Both may be empty for
// commit-only changes.
type Input struct {
	Labels     []string // raw label names, lowercased recommended but not required
	PRTitle    string
	PRBody     string
	CommitMsg  string // first line of the headline commit
	FilesPaths []string
}

// Categorize picks a ChangeType for the input. Order of precedence,
// highest wins:
//
//  1. Explicit label like "type:feat" / "feat" / "kind/bug"
//  2. A "changelog:" fenced block in the PR body, e.g. "changelog: fix"
//  3. Conventional Commit prefix in PRTitle or CommitMsg (feat:, fix!: …)
//  4. File-path heuristic (e.g. only docs/** → docs)
//  5. Default to "other"
//
// Confidence reflects how unambiguous the signal was. The renderer may
// hide low-confidence categories under a "Misc" bucket later, but the
// model keeps the analyzer's call.
func Categorize(in Input) CategoryDecision {
	if d, ok := fromLabels(in.Labels); ok {
		return d
	}
	if d, ok := fromChangelogBlock(in.PRBody); ok {
		return d
	}
	if d, ok := fromConventionalCommit(firstNonEmpty(in.PRTitle, in.CommitMsg)); ok {
		return d
	}
	if d, ok := fromFileHeuristic(in.FilesPaths); ok {
		return d
	}
	return CategoryDecision{Type: model.ChangeTypeOther, Confidence: 0.1, Source: "default"}
}

// labelMap is the closed translation from label-ish names to ChangeType.
// Order matters only for testing; lookups are by exact match after
// canonicalization (lowercase + strip "type:" / "kind/" prefix).
var labelMap = map[string]model.ChangeType{
	"feat":        model.ChangeTypeFeat,
	"feature":     model.ChangeTypeFeat,
	"enhancement": model.ChangeTypeFeat,
	"fix":         model.ChangeTypeFix,
	"bug":         model.ChangeTypeFix,
	"bugfix":      model.ChangeTypeFix,
	"breaking":    model.ChangeTypeBreaking,
	"deprecation": model.ChangeTypeDeprecation,
	"deprecated":  model.ChangeTypeDeprecation,
	"perf":        model.ChangeTypePerf,
	"performance": model.ChangeTypePerf,
	"security":    model.ChangeTypeSecurity,
	"docs":        model.ChangeTypeDocs,
	"doc":         model.ChangeTypeDocs,
	"chore":       model.ChangeTypeChore,
	"refactor":    model.ChangeTypeRefactor,
	"test":        model.ChangeTypeTest,
	"tests":       model.ChangeTypeTest,
	"build":       model.ChangeTypeBuild,
	"ci":          model.ChangeTypeCI,
}

func fromLabels(labels []string) (CategoryDecision, bool) {
	for _, raw := range labels {
		key := canonLabel(raw)
		if t, ok := labelMap[key]; ok {
			return CategoryDecision{Type: t, Confidence: 0.95, Source: "label"}, true
		}
	}
	return CategoryDecision{}, false
}

func canonLabel(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, p := range []string{"type:", "type/", "kind:", "kind/", "category:", "category/"} {
		if strings.HasPrefix(s, p) {
			s = strings.TrimPrefix(s, p)
			break
		}
	}
	return strings.TrimSpace(s)
}

// changelogBlockRe matches a single-line "changelog: <type>" hint
// anywhere in the PR body, case-insensitive.
var changelogBlockRe = regexp.MustCompile(`(?im)^\s*changelog\s*[:=]\s*([a-z]+)\s*$`)

func fromChangelogBlock(body string) (CategoryDecision, bool) {
	m := changelogBlockRe.FindStringSubmatch(body)
	if m == nil {
		return CategoryDecision{}, false
	}
	if t, ok := labelMap[strings.ToLower(m[1])]; ok {
		return CategoryDecision{Type: t, Confidence: 0.9, Source: "changelog-block"}, true
	}
	return CategoryDecision{}, false
}

// ccRe matches Conventional Commit prefixes: "feat", "fix", optional
// "(scope)", optional "!", then ":". Captures: 1=type, 2=scope, 3=bang.
var ccRe = regexp.MustCompile(`^([a-zA-Z]+)(?:\(([^)]+)\))?(!)?:\s+`)

func fromConventionalCommit(title string) (CategoryDecision, bool) {
	m := ccRe.FindStringSubmatch(strings.TrimSpace(title))
	if m == nil {
		return CategoryDecision{}, false
	}
	t, ok := labelMap[strings.ToLower(m[1])]
	if !ok {
		return CategoryDecision{}, false
	}
	conf := 0.85
	if m[3] == "!" {
		// Breaking marker; categorize.go records the type, breaking.go
		// flips the Breaking bit separately.
		conf = 0.9
	}
	return CategoryDecision{Type: t, Confidence: conf, Source: "cc-prefix"}, true
}

// fromFileHeuristic returns "docs" when every file is under docs/ or is
// a markdown file at the repo root; "test" when every file lives under
// a test directory. Otherwise no decision.
func fromFileHeuristic(paths []string) (CategoryDecision, bool) {
	if len(paths) == 0 {
		return CategoryDecision{}, false
	}
	allDocs, allTest := true, true
	for _, p := range paths {
		if !looksLikeDoc(p) {
			allDocs = false
		}
		if !looksLikeTest(p) {
			allTest = false
		}
	}
	switch {
	case allDocs:
		return CategoryDecision{Type: model.ChangeTypeDocs, Confidence: 0.6, Source: "heuristic"}, true
	case allTest:
		return CategoryDecision{Type: model.ChangeTypeTest, Confidence: 0.6, Source: "heuristic"}, true
	}
	return CategoryDecision{}, false
}

func looksLikeDoc(p string) bool {
	lp := strings.ToLower(p)
	if strings.HasPrefix(lp, "docs/") || strings.HasPrefix(lp, "doc/") {
		return true
	}
	return strings.HasSuffix(lp, ".md") || strings.HasSuffix(lp, ".rst") || strings.HasSuffix(lp, ".adoc")
}

func looksLikeTest(p string) bool {
	lp := strings.ToLower(p)
	if strings.Contains(lp, "/test/") || strings.Contains(lp, "/tests/") {
		return true
	}
	return strings.HasSuffix(lp, "_test.go") || strings.HasSuffix(lp, ".test.ts") || strings.HasSuffix(lp, ".test.js") || strings.HasSuffix(lp, ".spec.ts") || strings.HasSuffix(lp, ".spec.js")
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
