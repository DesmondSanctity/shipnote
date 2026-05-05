package repo_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DesmondSanctity/shipnote/internal/repo"
)

// requireGit skips the test if the git binary is unavailable on the host.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
}

func TestOpenRejectsNonRepo(t *testing.T) {
	t.Parallel()
	requireGit(t)
	dir := t.TempDir()
	if _, err := repo.Open(context.Background(), dir); err == nil {
		t.Fatal("expected error opening non-repo dir")
	}
}

// fixtureBaseEnv is the environment used for all git invocations during
// fixture construction. We pin author/committer identity here; dates
// are added per-commit by fixtureCommit.
func fixtureBaseEnv() []string {
	return append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
}

// fixtureRunGit runs git in dir with the supplied env. extraEnv extends
// (and can override) fixtureBaseEnv for a single call.
func fixtureRunGit(t *testing.T, dir string, extraEnv []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(fixtureBaseEnv(), extraEnv...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// fixtureCommit writes content to file inside dir, stages it, and
// commits with msg at base+offset minutes.
func fixtureCommit(t *testing.T, dir, msg, file, content string, base time.Time, offsetMin int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	dt := base.Add(time.Duration(offsetMin) * time.Minute).Format(time.RFC3339)
	extra := []string{"GIT_AUTHOR_DATE=" + dt, "GIT_COMMITTER_DATE=" + dt}
	fixtureRunGit(t, dir, nil, "add", file)
	fixtureRunGit(t, dir, extra, "commit", "-m", msg)
}
