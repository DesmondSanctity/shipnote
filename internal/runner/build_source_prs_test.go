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

// TestBuildSourcePRs_DedupesByPRNumber asserts that when the same PR
// is returned for multiple commits (e.g. a merge commit and the
// branch's original commit both resolve to the same PR via
// associatedPullRequests) we collapse them into a single SourcePR
// with all distinct commits attached.
func TestBuildSourcePRs_DedupesByPRNumber(t *testing.T) {
	t.Parallel()
	commits := []repo.Commit{
		{SHA: "merge", ShortSHA: "merge", Subject: "Merge pull request #7"},
		{SHA: "feat", ShortSHA: "feat", Subject: "feat: x"},
	}
	prs := []github.PR{
		{Number: 7, Title: "feat: x", MergeCommit: "merge"},
		{Number: 7, Title: "feat: x", MergeCommit: "merge"},
	}
	got := buildSourcePRs(prs, commits)
	if len(got) != 1 {
		t.Fatalf("expected 1 deduped SourcePR, got %d", len(got))
	}
	if got[0].PR.Number != 7 {
		t.Fatalf("expected PR #7, got #%d", got[0].PR.Number)
	}
	if len(got[0].Commits) != 1 || got[0].Commits[0].SHA != "merge" {
		t.Fatalf("expected only merge commit (no duplicate SHA), got %+v", got[0].Commits)
	}
}
