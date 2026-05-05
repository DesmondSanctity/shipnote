package runner_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DesmondSanctity/shipnote/internal/config"
	"github.com/DesmondSanctity/shipnote/internal/runner"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// fixtureRepo builds a small deterministic git repo: two commits before
// v0.1.0, one commit after. Returns the working-tree path.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	at := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	run := func(extra []string, args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(append([]string{}, env...), extra...)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	commit := func(msg, file, body string, off int) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, file), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		dt := at.Add(time.Duration(off) * time.Minute).Format(time.RFC3339)
		run(nil, "add", file)
		run([]string{"GIT_AUTHOR_DATE=" + dt, "GIT_COMMITTER_DATE=" + dt}, "commit", "-m", msg)
	}
	run(nil, "init", "-q", "-b", "main")
	run(nil, "config", "commit.gpgsign", "false")
	run(nil, "config", "tag.gpgsign", "false")
	run(nil, "remote", "add", "origin", "https://github.com/Acme/widgets.git")
	commit("feat: add readme", "README.md", "hello\n", 0)
	commit("fix: typo in readme", "README.md", "hello world\n", 10)
	tagDate := at.Add(15 * time.Minute).Format(time.RFC3339)
	run([]string{"GIT_AUTHOR_DATE=" + tagDate, "GIT_COMMITTER_DATE=" + tagDate},
		"tag", "-a", "v0.1.0", "-m", "release")
	commit("feat: add main.go", "main.go", "package main\n", 20)
	return dir
}

// TestRun_Deterministic is the v0.1 exit-gate test: the same inputs
// must produce byte-identical Markdown and JSON across runs.
func TestRun_Deterministic(t *testing.T) {
	t.Parallel()
	dir := fixtureRepo(t)
	in := runner.Inputs{
		RepoDir:   dir,
		ToRef:     "HEAD",
		Cfg:       config.Default(),
		NoNetwork: true,
	}
	a, err := runner.Run(context.Background(), in)
	if err != nil {
		t.Fatalf("Run a: %v", err)
	}
	b, err := runner.Run(context.Background(), in)
	if err != nil {
		t.Fatalf("Run b: %v", err)
	}
	if a.Markdown != b.Markdown {
		t.Fatalf("markdown differs across runs:\n--- a ---\n%s\n--- b ---\n%s", a.Markdown, b.Markdown)
	}
	if string(a.JSON) != string(b.JSON) {
		t.Fatalf("json differs across runs:\n--- a ---\n%s\n--- b ---\n%s", a.JSON, b.JSON)
	}
	if a.Release.Metadata.DeterminismKey == "" {
		t.Fatalf("expected non-empty determinism key")
	}
	if a.Release.Repo.Owner != "Acme" || a.Release.Repo.Name != "widgets" {
		t.Fatalf("expected owner/name from origin, got %+v", a.Release.Repo)
	}
}

// TestRun_GoldenMarkdown asserts the markdown rendering matches a
// checked-in snapshot. Update with `go test -run GoldenMarkdown -update`.
func TestRun_GoldenMarkdown(t *testing.T) {
	t.Parallel()
	dir := fixtureRepo(t)
	out, err := runner.Run(context.Background(), runner.Inputs{
		RepoDir:   dir,
		FromRef:   "v0.1.0",
		ToRef:     "HEAD",
		Cfg:       config.Default(),
		NoNetwork: true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := scrubVolatile(out.Markdown)
	goldenPath := filepath.Join("testdata", "golden_unreleased.md")
	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v (run with -update to create it)", err)
	}
	if got != string(want) {
		t.Fatalf("golden mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// scrubVolatile removes values that legitimately vary across machines
// (here: nothing yet, the runner is fully deterministic). Kept as a
// hook for future fields like host-derived paths.
func scrubVolatile(s string) string { return s }
