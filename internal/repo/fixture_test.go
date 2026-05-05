package repo_test

import (
	"testing"
	"time"
)

// fixture builds a small git repository on disk and returns its path.
// Layout: 3 commits on main, an annotated tag v0.1.0 on commit 2.
// All dates are pinned so tests are deterministic.
func fixture(t *testing.T) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	at := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	fixtureRunGit(t, dir, nil, "init", "-q", "-b", "main")
	fixtureRunGit(t, dir, nil, "config", "commit.gpgsign", "false")
	fixtureRunGit(t, dir, nil, "config", "tag.gpgsign", "false")

	fixtureCommit(t, dir, "feat: add readme", "README.md", "hello\n", at, 0)
	fixtureCommit(t, dir, "fix: typo in readme", "README.md", "hello world\n", at, 10)
	tagDate := at.Add(15 * time.Minute).Format(time.RFC3339)
	fixtureRunGit(t, dir,
		[]string{"GIT_AUTHOR_DATE=" + tagDate, "GIT_COMMITTER_DATE=" + tagDate},
		"tag", "-a", "v0.1.0", "-m", "release v0.1.0")
	fixtureCommit(t, dir, "feat: add main.go", "main.go", "package main\n", at, 20)
	return dir
}
