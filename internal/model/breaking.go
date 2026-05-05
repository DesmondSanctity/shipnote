package model

// BreakingChange is a denormalized record duplicating breaking-change data
// out of the owning Change for easier downstream consumption (CHANGELOG
// "Breaking" section, migration guides, audits).
//
// PackageID is empty when the change spans no specific package or affects
// multiple packages. ChangeID is always non-empty and back-references the
// parent Change in either Release.GlobalChanges or Package.Changes.
type BreakingChange struct {
	ChangeID     string    `json:"changeId"`
	PackageID    string    `json:"packageId,omitempty"`
	Summary      string    `json:"summary"`
	Impact       string    `json:"impact"`
	Migration    Migration `json:"migration"`
	IntroducedIn string    `json:"introducedIn"`
}
