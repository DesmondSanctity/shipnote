// Package model contains the normalized release schema types emitted
// and consumed across the collector → analyzer → renderer pipeline.
//
// These types are the inter-layer contract; behavior here is limited to
// constructors, validators, and canonical (de)serialization. Anything
// stateful or I/O-bound belongs in another package.
//
// The schema is documented in detail in docs/SCHEMA.md (and mirrored
// in schema/release.schema.json). Field names, JSON tags, and zero
// values must match that document — changing any of them is a schema
// change and requires an ADR plus a schemaVersion bump.
package model

// SchemaVersion is the version of the release schema this package emits.
// Bumped only via an ADR. Stays in sync with internal/version.SchemaVersion.
const SchemaVersion = "1.0.0"

// Release is the top-level normalized release record.
type Release struct {
	SchemaVersion string `json:"schemaVersion"`

	ID                string `json:"id"`
	PreviousReleaseID string `json:"previousReleaseId,omitempty"`

	Metadata Metadata `json:"metadata"`
	Repo     Repo     `json:"repo"`
	Range    Range    `json:"range"`

	Release ReleaseHeader `json:"release"`

	Summary Summary `json:"summary"`
	Stats   Stats   `json:"stats"`

	Packages        []Package        `json:"packages"`
	GlobalChanges   []Change         `json:"globalChanges"`
	BreakingChanges []BreakingChange `json:"breakingChanges"`
	Contributors    []Contributor    `json:"contributors"`
	Excluded        []ExcludedRecord `json:"excluded"`
}
