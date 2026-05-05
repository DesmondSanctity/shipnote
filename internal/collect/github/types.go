package github

// MergeStrategy classifies how a PR ended up on its base branch.
// Mirrors the model.MergeStrategy enum but lives here as the wire-level
// value the GraphQL API returns; the analyzer translates between them.
type MergeStrategy string

// Allowed MergeStrategy values.
const (
	MergeStrategySquash  MergeStrategy = "squash"
	MergeStrategyMerge   MergeStrategy = "merge"
	MergeStrategyRebase  MergeStrategy = "rebase"
	MergeStrategyUnknown MergeStrategy = "unknown"
)

// User is a GitHub account (human or bot).
type User struct {
	Login     string
	Name      string
	URL       string
	AvatarURL string
	IsBot     bool
}

// Label is a PR label. Color is the 6-char hex without the leading "#".
type Label struct {
	Name  string
	Color string
}

// File is one file changed by a PR.
type File struct {
	Path      string
	Additions int
	Deletions int
	Status    string // added, modified, removed, renamed, copied, changed
}

// PR is a normalized pull-request record. Field set is intentionally
// the union of what the analyzer needs; we do not pass through every
// GraphQL field.
type PR struct {
	Number        int
	Title         string
	Body          string
	URL           string
	State         string // OPEN, CLOSED, MERGED
	MergedAt      string // RFC3339 UTC; empty if not merged
	MergeCommit   string // SHA of the merge commit; empty if squashed without one
	BaseRef       string
	HeadRef       string
	Author        User
	Labels        []Label
	Files         []File
	MergeStrategy MergeStrategy
}
