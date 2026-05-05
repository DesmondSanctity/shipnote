package repo

import (
	"context"
	"fmt"
	"strings"
)

// Tag is a single git tag. Date is the commit date of the tagged commit
// (NOT the tagger date on annotated tags) so it sorts consistently with
// the commit graph and is determinism-friendly.
type Tag struct {
	Name string
	SHA  string
	Date string // RFC3339 UTC
}

// Tags returns all tags whose name matches the given glob (e.g. "v*").
// Empty pattern returns every tag. Results are ordered by commit date
// ascending; ties are broken by tag name to keep ordering stable.
func (g *Git) Tags(ctx context.Context, pattern string) ([]Tag, error) {
	args := []string{
		"for-each-ref",
		"--sort=creatordate",
		"--format=%(refname:short)\x1f%(objectname)\x1f%(creatordate:iso-strict)",
	}
	if pattern == "" {
		args = append(args, "refs/tags/")
	} else {
		args = append(args, "refs/tags/"+pattern)
	}
	lines, err := g.runLines(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	tags := make([]Tag, 0, len(lines))
	for _, ln := range lines {
		parts := strings.SplitN(ln, "\x1f", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("malformed tag line: %q", ln)
		}
		tags = append(tags, Tag{
			Name: parts[0],
			SHA:  parts[1],
			Date: parts[2],
		})
	}
	return tags, nil
}

// LatestTag returns the most-recently-created tag matching pattern, or
// the zero value (with ok=false) if none exist.
func (g *Git) LatestTag(ctx context.Context, pattern string) (Tag, bool, error) {
	tags, err := g.Tags(ctx, pattern)
	if err != nil {
		return Tag{}, false, err
	}
	if len(tags) == 0 {
		return Tag{}, false, nil
	}
	return tags[len(tags)-1], true, nil
}
