package github_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/collect/github"
)

func TestPRsForCommitsHappyPath(t *testing.T) {
	t.Parallel()
	fs := newFakeServer(t, happyPathResponse)
	c := github.New(github.Options{Endpoint: fs.URL, Token: "test-token"})

	prs, err := c.PRsForCommits(context.Background(), "acme", "sdk", []string{"aaa", "bbb"})
	if err != nil {
		t.Fatalf("PRsForCommits: %v", err)
	}
	if len(prs) != 2 {
		t.Fatalf("len = %d, want 2", len(prs))
	}
	if prs[0].Number != 42 || prs[0].Title != "feat: x" || prs[0].MergeCommit != "merge1" {
		t.Errorf("pr[0] wrong: %+v", prs[0])
	}
	if prs[0].Author.Login != "maya" || prs[0].Author.IsBot {
		t.Errorf("author wrong: %+v", prs[0].Author)
	}
	if len(prs[0].Labels) != 1 || prs[0].Labels[0].Name != "feat" {
		t.Errorf("labels wrong: %+v", prs[0].Labels)
	}
	if len(prs[0].Files) != 1 || prs[0].Files[0].Status != "modified" {
		t.Errorf("files wrong: %+v", prs[0].Files)
	}
	if prs[1].Number != 0 {
		t.Errorf("pr[1] should be empty for unmatched commit, got %+v", prs[1])
	}

	if got := fs.lastHeaders.Get("Authorization"); got != "Bearer test-token" {
		t.Errorf("auth header = %q", got)
	}
	if !strings.Contains(fs.lastHeaders.Get("User-Agent"), "shipnote") {
		t.Errorf("UA missing shipnote tag: %q", fs.lastHeaders.Get("User-Agent"))
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(fs.lastBody), &sent); err != nil {
		t.Fatalf("decode sent body: %v", err)
	}
	q, _ := sent["query"].(string)
	if !strings.Contains(q, "c0:") || !strings.Contains(q, "c1:") {
		t.Errorf("query missing aliases: %q", q)
	}
}
