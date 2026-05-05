package analyze

import "testing"

func TestStripCCPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"feat: add thing", "add thing"},
		{"fix(runner): skip empty PR slots", "skip empty PR slots"},
		{"feat!: drop legacy api", "drop legacy api"},
		{"chore(release): cut v0.1.0", "cut v0.1.0"},
		{"random unstructured title", "random unstructured title"},
		{"   feat: padded   ", "padded"},
		{"", ""},
		{"WIP: experiment", "experiment"}, // matches lenient regex; harmless
	}
	for _, tc := range cases {
		if got := StripCCPrefix(tc.in); got != tc.want {
			t.Errorf("StripCCPrefix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
