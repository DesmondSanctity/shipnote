package runner

import (
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/collect/github"
	"github.com/DesmondSanctity/shipnote/internal/repo"
)

// TestBuildSourcePRs_SkipsEmpty asserts that PR slots returned by
// PRsForCommits for direct-push commits (Number == 0) are not turned
// into empty SourcePR records the analyzer would emit as blank rows.
func TestBuildSourcePRs_SkipsEmpty(t *testing.T) {
	t.Parallel()
	commits := []repo.Commit{{SHA: "aaa", ShortSHA: "aaa", Subject: "x"}}
	prs := []github.PR{
		{Number: 0}, // direct push, no PR
		{Number: 7, Title: "feat: x", MergeCommit: "aaa"}, // real PR
		{Number: 0}, // another direct push
	}
	got := buildSourcePRs(prs, commits)
	if len(got) != 1 {
		t.Fatalf("expected 1 SourcePR, got %d", len(got))
	}
	if got[0].PR.Number != 7 {
		t.Fatalf("expected PR #7, got #%d", got[0].PR.Number)
	}
	if len(got[0].Commits) != 1 || got[0].Commits[0].SHA != "aaa" {
		t.Fatalf("expected commit aaa attached, got %+v", got[0].Commits)
	}
}
