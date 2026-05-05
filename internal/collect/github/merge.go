package github

import "strings"

// detectMergeStrategy classifies a PR's landing commit by parent count.
//
// GitHub's three merge modes leave distinct shapes on the base branch:
//
//   - "Create a merge commit" produces a non-fast-forward merge with
//     two parents (base tip + PR head). parentCount == 2.
//   - "Squash and merge" collapses the PR into a single new commit on
//     base whose only parent is the previous base tip. parentCount == 1.
//   - "Rebase and merge" replays each PR commit linearly on base; the
//     mergeCommit GraphQL field points at the last replayed commit and
//     also has one parent. parentCount == 1.
//
// Squash and rebase share parentCount == 1, so we cannot disambiguate
// without inspecting the PR commit list. The analyzer treats both as
// "linear landings"; if a caller needs the exact mode it can override
// using local git data later.
//
// state is the PR's GraphQL state ("MERGED", "CLOSED", "OPEN"). For
// non-merged PRs we always return Unknown.
func detectMergeStrategy(parentCount int, state string) MergeStrategy {
	if !strings.EqualFold(state, "MERGED") {
		return MergeStrategyUnknown
	}
	switch {
	case parentCount >= 2:
		return MergeStrategyMerge
	case parentCount == 1:
		// Default to squash since it is the modern GitHub default and
		// matches most repos; rebase looks identical at this level.
		return MergeStrategySquash
	default:
		return MergeStrategyUnknown
	}
}
