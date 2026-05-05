package github

import "testing"

func TestDetectMergeStrategy(t *testing.T) {
	cases := []struct {
		name    string
		parents int
		state   string
		want    MergeStrategy
	}{
		{"merge commit", 2, "MERGED", MergeStrategyMerge},
		{"three parent octopus", 3, "MERGED", MergeStrategyMerge},
		{"squash or rebase", 1, "MERGED", MergeStrategySquash},
		{"closed not merged", 1, "CLOSED", MergeStrategyUnknown},
		{"open", 1, "OPEN", MergeStrategyUnknown},
		{"merged but no parents", 0, "MERGED", MergeStrategyUnknown},
		{"case-insensitive state", 2, "merged", MergeStrategyMerge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := detectMergeStrategy(tc.parents, tc.state)
			if got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}
