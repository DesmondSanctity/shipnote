package model

// Metadata captures generation context. Excluded from determinism checks.
// GeneratedAt MUST equal Range.To.Date — never wall-clock time.
type Metadata struct {
	GeneratedAt    string    `json:"generatedAt"`
	Generator      Generator `json:"generator"`
	DeterminismKey string    `json:"determinismKey"`
	Sources        []string  `json:"sources"`
}

// Generator identifies the tool and version that produced the record.
type Generator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Repo identifies the source repository.
type Repo struct {
	Provider      string `json:"provider"`
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	DefaultBranch string `json:"defaultBranch"`
}

// Range is the commit range covered by this release record.
type Range struct {
	From RangePoint `json:"from"`
	To   RangePoint `json:"to"`
}

// RangePoint identifies one end of a range.
type RangePoint struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Date string `json:"date"`
}

// ReleaseHeader carries publication state. The JSON key is "release";
// the Go type is named differently to avoid colliding with the outer Release.
type ReleaseHeader struct {
	State        ReleaseState `json:"state"`
	Name         string       `json:"name"`
	Tag          string       `json:"tag"`
	IsPrerelease bool         `json:"isPrerelease"`
	PublishedAt  *string      `json:"publishedAt"`
}

// ReleaseState enumerates the lifecycle stages a release record can be in.
type ReleaseState string

// Allowed ReleaseState values.
const (
	ReleaseStateUnreleased ReleaseState = "unreleased"
	ReleaseStateDraft      ReleaseState = "draft"
	ReleaseStateReleased   ReleaseState = "released"
)
