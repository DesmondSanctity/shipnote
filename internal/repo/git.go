// Package repo is a thin, well-typed wrapper over the local git binary.
// All of shipnote's git access flows through here so the rest of the
// codebase can stay free of os/exec calls and string parsing of git
// output. We intentionally do NOT embed go-git or libgit2 — the tradeoff
// is a hard runtime dependency on /usr/bin/git, in exchange for matching
// every git config / hook / signature / replace-ref the user already has.
package repo

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Git is a handle to a local git working tree or bare repository.
// All methods are safe to call concurrently. Construct with Open.
type Git struct {
	dir    string
	gitBin string
}

// Option configures a Git handle.
type Option func(*Git)

// WithGitBinary overrides the path to the git executable. Defaults to
// looking up "git" on $PATH.
func WithGitBinary(path string) Option {
	return func(g *Git) { g.gitBin = path }
}

// Open returns a Git handle rooted at dir. dir must point to a working
// tree or a bare repository; this is verified eagerly with `git rev-parse`.
func Open(ctx context.Context, dir string, opts ...Option) (*Git, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("repo.Open: dir is empty")
	}
	g := &Git{dir: dir, gitBin: "git"}
	for _, opt := range opts {
		opt(g)
	}
	if _, err := exec.LookPath(g.gitBin); err != nil {
		return nil, fmt.Errorf("git binary not found: %w", err)
	}
	if _, err := g.run(ctx, "rev-parse", "--git-dir"); err != nil {
		return nil, fmt.Errorf("not a git repository at %q: %w", dir, err)
	}
	return g, nil
}

// Dir reports the working-tree (or bare-repo) directory the handle is
// bound to.
func (g *Git) Dir() string { return g.dir }
