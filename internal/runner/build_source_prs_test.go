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
	got := buildSourcePRs(prs, commits, false, nil)
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
	got := buildSourcePRs(prs, commits, false, nil)
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

// TestBuildSourcePRs_IncludesDirectPushes asserts that when direct
// pushes are enabled, commits on the first-parent line that no PR
// claims are emitted as PR-less (Number == 0) entries with author
// + commit metadata copied from the commit.
func TestBuildSourcePRs_IncludesDirectPushes(t *testing.T) {
	t.Parallel()
	commits := []repo.Commit{
		{SHA: "merge", ShortSHA: "merge", Subject: "Merge pull request #7"},
		{
			SHA: "direct", ShortSHA: "direct", Subject: "fix: hotfix on main",
			AuthorName: "Alice", AuthorEmail: "alice@example.com",
		},
	}
	prs := []github.PR{
		{Number: 7, Title: "feat: x", MergeCommit: "merge"},
		{Number: 0}, // direct push commit, no PR
	}
	firstParents := map[string]struct{}{
		"merge":  {},
		"direct": {},
	}
	got := buildSourcePRs(prs, commits, true, firstParents)
	if len(got) != 2 {
		t.Fatalf("expected 2 SourcePRs (PR #7 + direct), got %d", len(got))
	}
	// PR-backed entry comes first (insertion order preserved).
	if got[0].PR.Number != 7 {
		t.Fatalf("expected first entry PR #7, got #%d", got[0].PR.Number)
	}
	if got[1].PR.Number != 0 {
		t.Fatalf("expected second entry to be direct push (Number==0), got %+v", got[1].PR)
	}
	if got[1].PR.Title != "fix: hotfix on main" {
		t.Fatalf("expected commit subject as title, got %q", got[1].PR.Title)
	}
	if got[1].PR.Author.Name != "Alice" {
		t.Fatalf("expected author Alice, got %q", got[1].PR.Author.Name)
	}
	if len(got[1].Commits) != 1 || got[1].Commits[0].SHA != "direct" {
		t.Fatalf("expected commit SHA direct attached, got %+v", got[1].Commits)
	}
}

// TestBuildSourcePRs_DirectPushesRespectFirstParent asserts that direct
// pushes outside the first-parent line are dropped (they belong to
// merged feature branches that PR association already covered).
func TestBuildSourcePRs_DirectPushesRespectFirstParent(t *testing.T) {
	t.Parallel()
	commits := []repo.Commit{
		{SHA: "branch-internal", ShortSHA: "branch", Subject: "wip"},
	}
	prs := []github.PR{{Number: 0}}
	firstParents := map[string]struct{}{} // empty: this commit is not on first-parent
	got := buildSourcePRs(prs, commits, true, firstParents)
	if len(got) != 0 {
		t.Fatalf("expected commit off first-parent to be dropped, got %d", len(got))
	}
}
