package repo

import "testing"

func TestParseGitHubRemote(t *testing.T) {
	cases := []struct {
		in          string
		owner, name string
		ok          bool
	}{
		{"https://github.com/Acme/widgets.git", "Acme", "widgets", true},
		{"https://github.com/Acme/widgets", "Acme", "widgets", true},
		{"git@github.com:Acme/widgets.git", "Acme", "widgets", true},
		{"ssh://git@github.com/Acme/widgets.git", "Acme", "widgets", true},
		{"https://gitlab.com/Acme/widgets.git", "", "", false},
		{"", "", "", false},
	}
	for _, tc := range cases {
		o, n, ok := ParseGitHubRemote(tc.in)
		if ok != tc.ok || o != tc.owner || n != tc.name {
			t.Errorf("ParseGitHubRemote(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tc.in, o, n, ok, tc.owner, tc.name, tc.ok)
		}
	}
}
