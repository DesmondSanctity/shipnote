package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitignoreEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// new file
	if err := ensureGitignoreEntry(dir, ".shipnote/"); err != nil {
		t.Fatalf("first call: %v", err)
	}
	body := mustRead(t, filepath.Join(dir, ".gitignore"))
	if body != ".shipnote/\n" {
		t.Fatalf("new file body: %q", body)
	}

	// idempotent
	if err := ensureGitignoreEntry(dir, ".shipnote/"); err != nil {
		t.Fatalf("second call: %v", err)
	}
	body2 := mustRead(t, filepath.Join(dir, ".gitignore"))
	if body2 != body {
		t.Fatalf("not idempotent: before=%q after=%q", body, body2)
	}

	// append to existing without trailing newline
	dir2 := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir2, ".gitignore"), []byte("node_modules/"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureGitignoreEntry(dir2, ".shipnote/"); err != nil {
		t.Fatalf("append: %v", err)
	}
	got := mustRead(t, filepath.Join(dir2, ".gitignore"))
	if !strings.Contains(got, "node_modules/") || !strings.Contains(got, ".shipnote/") {
		t.Fatalf("appended body wrong: %q", got)
	}
}

func TestAppendPRTemplate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := appendPRTemplate(dir); err != nil {
		t.Fatalf("first: %v", err)
	}
	path := filepath.Join(dir, ".github", "pull_request_template.md")
	body := mustRead(t, path)
	if !strings.Contains(body, "<!-- shipnote -->") {
		t.Fatalf("missing marker: %q", body)
	}
	// idempotent
	if err := appendPRTemplate(dir); err != nil {
		t.Fatalf("second: %v", err)
	}
	body2 := mustRead(t, path)
	if body2 != body {
		t.Fatalf("not idempotent")
	}
}

func mustRead(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}
