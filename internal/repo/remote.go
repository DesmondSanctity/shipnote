package repo

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// OriginURL returns the URL of the "origin" remote, or any remote when
// there is exactly one. Empty string + nil error if no remote is set.
func (g *Git) OriginURL(ctx context.Context) (string, error) {
	out, err := g.run(ctx, "remote", "get-url", "origin")
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	// origin missing — fall back to the only remote, if there is one.
	lines, lerr := g.runLines(ctx, "remote")
	if lerr != nil {
		return "", fmt.Errorf("list remotes: %w", lerr)
	}
	if len(lines) == 1 {
		out, err := g.run(ctx, "remote", "get-url", lines[0])
		if err != nil {
			return "", fmt.Errorf("get remote %s url: %w", lines[0], err)
		}
		return strings.TrimSpace(string(out)), nil
	}
	return "", nil
}

// ownerNameRe matches the {owner}/{name} segment from the common
// GitHub URL forms:
//   - https://github.com/owner/name(.git)?
//   - git@github.com:owner/name(.git)?
//   - ssh://git@github.com/owner/name(.git)?
var ownerNameRe = regexp.MustCompile(`(?:github\.com[/:])([^/]+)/([^/]+?)(?:\.git)?/?$`)

// ParseGitHubRemote extracts (owner, name) from a GitHub remote URL.
// Returns ok=false when the URL does not look like a GitHub remote.
func ParseGitHubRemote(rawURL string) (owner, name string, ok bool) {
	m := ownerNameRe.FindStringSubmatch(strings.TrimSpace(rawURL))
	if m == nil {
		return "", "", false
	}
	return m[1], strings.TrimSuffix(m[2], ".git"), true
}
