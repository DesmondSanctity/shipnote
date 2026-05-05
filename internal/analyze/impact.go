package analyze

import "strings"

// PackageMapper resolves a file path to a package ID. Single-package
// repos always return "" (the root package); multi-package support
// arrives in v0.2 with workspace detection.
type PackageMapper interface {
	PackageFor(path string) string
}

// SinglePackageMapper maps every file to the configured root package
// ID. The zero value uses "" which the renderer treats as "root".
type SinglePackageMapper struct {
	RootID string
}

// PackageFor implements PackageMapper.
func (m SinglePackageMapper) PackageFor(string) string { return m.RootID }

// PackagesForFiles returns the deduplicated, sorted set of package IDs
// that the given file paths map to. An empty input yields nil.
func PackagesForFiles(mapper PackageMapper, paths []string) []string {
	if len(paths) == 0 || mapper == nil {
		return nil
	}
	seen := make(map[string]struct{}, 4)
	for _, p := range paths {
		id := strings.TrimSpace(mapper.PackageFor(p))
		seen[id] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	// stable: empty string sorts first which makes "root" predictable
	sortStrings(out)
	return out
}

func sortStrings(xs []string) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
}
