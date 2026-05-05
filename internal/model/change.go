package model

// Change is the atomic unit of a release. Always traceable back to its source.
// Audience is intentionally NOT a field on Change — it is a render-time concern.
type Change struct {
	ID          string     `json:"id"`
	Type        ChangeType `json:"type"`
	Scope       string     `json:"scope,omitempty"`
	Breaking    bool       `json:"breaking"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`

	Source ChangeSource `json:"source"`

	Author    Contributor   `json:"author"`
	CoAuthors []Contributor `json:"coAuthors"`

	Packages []string `json:"packages"`

	Files FileSummary `json:"files"`

	References References `json:"references"`

	Migration   *Migration  `json:"migration"`
	AISummaries AISummaries `json:"aiSummaries"`

	Confidence Confidence `json:"confidence"`
}

// ChangeType is the analyzer-assigned category for a change.
type ChangeType string

// Allowed ChangeType values. Mirrors the closed category set documented in
// docs/SCHEMA.md and used by the markdown renderer's locked ordering.
const (
	ChangeTypeFeat        ChangeType = "feat"
	ChangeTypeFix         ChangeType = "fix"
	ChangeTypeBreaking    ChangeType = "breaking"
	ChangeTypeDeprecation ChangeType = "deprecation"
	ChangeTypePerf        ChangeType = "perf"
	ChangeTypeSecurity    ChangeType = "security"
	ChangeTypeDocs        ChangeType = "docs"
	ChangeTypeChore       ChangeType = "chore"
	ChangeTypeRefactor    ChangeType = "refactor"
	ChangeTypeTest        ChangeType = "test"
	ChangeTypeBuild       ChangeType = "build"
	ChangeTypeCI          ChangeType = "ci"
	ChangeTypeOther       ChangeType = "other"
)

// ChangeSource describes how a change entered the system.
type ChangeSource struct {
	Kind    SourceKind `json:"kind"`
	Adapter string     `json:"adapter"`
	PR      *PRSource  `json:"pr"`
	Commits []Commit   `json:"commits"`
}

// SourceKind enumerates change-origin kinds.
type SourceKind string

// Allowed SourceKind values.
const (
	SourceKindPR            SourceKind = "pr"
	SourceKindCommit        SourceKind = "commit"
	SourceKindChangeset     SourceKind = "changeset"
	SourceKindReleasePlease SourceKind = "release-please"
)

// PRSource carries pull-request metadata for a change.
type PRSource struct {
	Number        int           `json:"number"`
	URL           string        `json:"url"`
	MergedAt      string        `json:"mergedAt"`
	MergeStrategy MergeStrategy `json:"mergeStrategy"`
	Labels        []string      `json:"labels"`
}

// MergeStrategy identifies how a PR was merged.
type MergeStrategy string

// Allowed MergeStrategy values.
const (
	MergeSquash  MergeStrategy = "squash"
	MergeMerge   MergeStrategy = "merge"
	MergeRebase  MergeStrategy = "rebase"
	MergeUnknown MergeStrategy = "unknown"
)

// Commit is a single commit linked to a change.
type Commit struct {
	SHA      string `json:"sha"`
	ShortSHA string `json:"shortSha"`
	URL      string `json:"url,omitempty"`
	Message  string `json:"message"`
}

// FileSummary summarizes the file footprint of a change. The full per-file
// list is populated only when the user passes --full-files; otherwise List
// stays nil and Top contains the largest-diff entries.
type FileSummary struct {
	Count     int        `json:"count"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	Top       []FileDiff `json:"top"`
	List      []FileDiff `json:"list"` // nil unless --full-files
}

// FileDiff describes one file's change footprint.
type FileDiff struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Status    string `json:"status"` // added | modified | removed | renamed
}

// References links a change to issues, sibling PRs, and external URLs.
type References struct {
	Closes    []string `json:"closes"`
	LinkedPRs []string `json:"linkedPrs"`
	External  []string `json:"external"`
}

// Migration is an optional structured migration block attached to a change.
// For breaking changes the analyzer typically populates it; the BreakingChange
// record references the same struct shape.
type Migration struct {
	Summary   string `json:"summary"`
	Before    string `json:"before,omitempty"`
	After     string `json:"after,omitempty"`
	Automated bool   `json:"automated"`
}

// Confidence is the analyzer's self-assessment of its decisions.
type Confidence struct {
	Category       float64 `json:"category"`
	Breaking       float64 `json:"breaking"`
	PackageMapping float64 `json:"packageMapping"`
}
