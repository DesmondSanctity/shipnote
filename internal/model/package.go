package model

// Package is one logical unit (module, app, library) within the repo.
// For non-monorepo repos there is exactly one Package with ID "root".
type Package struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Path            string           `json:"path"`
	Ecosystem       string           `json:"ecosystem"`
	TagPrefix       string           `json:"tagPrefix"`
	Version         PackageVersion   `json:"version"`
	Summary         Summary          `json:"summary"`
	Changes         []Change         `json:"changes"`
	BreakingChanges []BreakingChange `json:"breakingChanges"`
	Stats           Stats            `json:"stats"`
}

// PackageVersion captures the previous and proposed/current version plus
// the bump kind implied by the changes. Previous may be empty for a
// package's first release.
type PackageVersion struct {
	Previous string      `json:"previous"`
	Current  string      `json:"current"`
	Bump     PackageBump `json:"bump"`
}

// PackageBump enumerates the kind of version bump implied by the changes.
type PackageBump string

// Allowed PackageBump values.
const (
	PackageBumpNone  PackageBump = "none"
	PackageBumpPatch PackageBump = "patch"
	PackageBumpMinor PackageBump = "minor"
	PackageBumpMajor PackageBump = "major"
	PackageBumpPre   PackageBump = "prerelease"
)
