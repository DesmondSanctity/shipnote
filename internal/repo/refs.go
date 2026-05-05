package repo

import (
	"context"
	"fmt"
	"strings"
)

// ResolveRef returns the full SHA + commit date for a ref (tag, branch,
// short SHA, "HEAD", etc.). Date is RFC3339 UTC.
func (g *Git) ResolveRef(ctx context.Context, ref string) (sha, date string, err error) {
	out, err := g.run(ctx,
		"log", "-1",
		"--format=%H\x1f%cI",
		ref, "--",
	)
	if err != nil {
		return "", "", fmt.Errorf("resolve ref %q: %w", ref, err)
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), "\x1f", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed log output for %q: %q", ref, out)
	}
	return parts[0], parts[1], nil
}

// MergeBase returns the best common ancestor SHA of two refs. Used when
// the "from" tag is on a release branch that has since diverged from
// the default branch — the analyzer wants the merge base, not the tag
// itself, so feature commits already merged into the branch don't get
// double-counted.
func (g *Git) MergeBase(ctx context.Context, a, b string) (string, error) {
	out, err := g.run(ctx, "merge-base", a, b)
	if err != nil {
		return "", fmt.Errorf("merge-base %s %s: %w", a, b, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CurrentBranch returns the short name of the branch HEAD currently
// points to (e.g. "main", "release/2.4"). Returns "HEAD" when HEAD is
// detached. Empty string + nil error when the repo has no commits yet.
func (g *Git) CurrentBranch(ctx context.Context) (string, error) {
	out, err := g.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// DefaultBranch returns the repository's default branch name. It first
// consults `refs/remotes/origin/HEAD` (the value `git clone` records),
// then falls back to the local `init.defaultBranch` config, and finally
// to "main" when neither is set. Never returns an error for a missing
// remote — callers treat this as best-effort metadata.
func (g *Git) DefaultBranch(ctx context.Context) string {
	if out, err := g.run(ctx, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); err == nil {
		s := strings.TrimSpace(string(out))
		if i := strings.LastIndex(s, "/"); i >= 0 && i < len(s)-1 {
			return s[i+1:]
		}
		if s != "" {
			return s
		}
	}
	if out, err := g.run(ctx, "config", "--get", "init.defaultBranch"); err == nil {
		if s := strings.TrimSpace(string(out)); s != "" {
			return s
		}
	}
	return "main"
}

// FirstCommit returns the SHA of the repository's root commit (the
// commit reachable from HEAD with no parents). Used as a fallback when
// the user asks for "everything since the beginning of the repo".
func (g *Git) FirstCommit(ctx context.Context) (string, error) {
	out, err := g.run(ctx, "rev-list", "--max-parents=0", "HEAD")
	if err != nil {
		return "", fmt.Errorf("find first commit: %w", err)
	}
	first := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	if first == "" {
		return "", fmt.Errorf("repository has no commits")
	}
	return first, nil
}
