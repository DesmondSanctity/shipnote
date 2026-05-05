package model

// Contributor is a deduplicated author/co-author across the range.
// Login is empty when the contributor could not be mapped to a provider account.
// Changes counts the number of Change records this contributor authored or co-authored.
type Contributor struct {
	Login   string `json:"login"`
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	Changes int    `json:"changes"`
}

// ExcludedRecord captures a source item intentionally dropped by the
// analyzer (chore-only PRs, bot noise, merges) along with a stable
// reason code. ID is the source-system identifier (PR number, commit
// SHA, etc.) prefixed with the adapter name.
type ExcludedRecord struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}
