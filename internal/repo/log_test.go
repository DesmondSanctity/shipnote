package repo_test

import (
	"context"
	"strings"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/repo"
)

func TestResolveRangeDefaults(t *testing.T) {
	t.Parallel()
	dir := fixture(t)
	g, err := repo.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// Default: from = latest v* tag, to = HEAD.
	rg, err := g.ResolveRange(context.Background(), "", "", "v*")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if rg.From.Ref != "v0.1.0" {
		t.Errorf("from.ref = %q, want v0.1.0", rg.From.Ref)
	}
	if rg.To.Ref != "HEAD" {
		t.Errorf("to.ref = %q, want HEAD", rg.To.Ref)
	}
	if !strings.HasPrefix(rg.From.Date, "2026-04-01T") {
		t.Errorf("from.date unexpected: %q", rg.From.Date)
	}
}

func TestLogInRange(t *testing.T) {
	t.Parallel()
	dir := fixture(t)
	g, err := repo.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	commits, err := g.Log(context.Background(), "v0.1.0", "HEAD")
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("want 1 commit since v0.1.0, got %d: %+v", len(commits), commits)
	}
	c := commits[0]
	if c.Subject != "feat: add main.go" {
		t.Errorf("subject = %q", c.Subject)
	}
	if c.AuthorEmail != "test@example.com" {
		t.Errorf("author = %q", c.AuthorEmail)
	}
	if c.SHA == "" || len(c.ShortSHA) < 7 {
		t.Errorf("missing SHA fields: %+v", c)
	}
}

func TestCommitFiles(t *testing.T) {
	t.Parallel()
	dir := fixture(t)
	g, err := repo.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	commits, err := g.Log(context.Background(), "", "HEAD")
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	// HEAD commit added main.go
	files, err := g.CommitFiles(context.Background(), commits[0].SHA)
	if err != nil {
		t.Fatalf("files: %v", err)
	}
	if len(files) != 1 || files[0].Path != "main.go" {
		t.Fatalf("unexpected files: %+v", files)
	}
	if files[0].Status != "A" {
		t.Errorf("status = %q, want A", files[0].Status)
	}
	if files[0].Additions != 1 {
		t.Errorf("additions = %d, want 1", files[0].Additions)
	}
}
