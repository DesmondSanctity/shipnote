// Package version exposes build-time identification information.
//
// Values are populated via -ldflags at build time by GoReleaser and the
// Makefile. Runtime code must not mutate them.
package version

// Build-time variables. Override with:
//
//	go build -ldflags "-X github.com/DesmondSanctity/shipnote/internal/version.Version=v0.1.0 \
//	                   -X github.com/DesmondSanctity/shipnote/internal/version.Commit=$(git rev-parse HEAD) \
//	                   -X github.com/DesmondSanctity/shipnote/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// SchemaVersion is the release-schema version this binary emits.
// Bump only via an ADR.
const SchemaVersion = "1.0.0"
