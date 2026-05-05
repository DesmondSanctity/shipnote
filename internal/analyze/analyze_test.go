package analyze

import (
	"testing"

	gh "github.com/DesmondSanctity/shipnote/internal/collect/github"
	"github.com/DesmondSanctity/shipnote/internal/model"
)

func mkPR(num int, title string, labels []string, author gh.User, files []gh.File) gh.PR {
	ls := make([]gh.Label, len(labels))
	for i, n := range labels {
		ls[i] = gh.Label{Name: n}
	}
	return gh.PR{
		Number: num,
		Title:  title,
		URL:    "https://example.com/pr/" + title,
		Author: author,
		Labels: ls,
		Files:  files,
	}
}

func TestAnalyze_Empty(t *testing.T) {
	rel := Analyze(nil, Options{RootPkg: "root"})
	if rel.SchemaVersion != model.SchemaVersion {
		t.Fatalf("schema version: %s", rel.SchemaVersion)
	}
	if len(rel.GlobalChanges) != 0 || len(rel.Contributors) != 0 {
		t.Fatalf("expected empty release, got %#v", rel)
	}
}

func TestAnalyze_BasicFeat(t *testing.T) {
	prs := []SourcePR{{
		PR: mkPR(1, "feat: add login", []string{"feature"},
			gh.User{Login: "alice", Name: "Alice"},
			[]gh.File{{Path: "src/login.go", Additions: 10, Deletions: 1, Status: "added"}}),
	}}
	rel := Analyze(prs, Options{RootPkg: "root"})
	if len(rel.GlobalChanges) != 1 {
		t.Fatalf("want 1 change, got %d", len(rel.GlobalChanges))
	}
	ch := rel.GlobalChanges[0]
	if ch.Type != model.ChangeTypeFeat {
		t.Fatalf("want feat, got %s", ch.Type)
	}
	if ch.Breaking {
		t.Fatalf("should not be breaking")
	}
	if ch.ID != "github-prs:1" {
		t.Fatalf("id: %s", ch.ID)
	}
	if len(rel.Contributors) != 1 || rel.Contributors[0].Login != "alice" {
		t.Fatalf("contribs: %#v", rel.Contributors)
	}
	if rel.Stats.Contributors != 1 {
		t.Fatalf("stats.contributors: %d", rel.Stats.Contributors)
	}
}

func TestAnalyze_BreakingFloats(t *testing.T) {
	prs := []SourcePR{
		{PR: mkPR(1, "fix: small fix", nil, gh.User{Login: "a"}, nil)},
		{PR: func() gh.PR {
			p := mkPR(2, "feat!: drop legacy api", nil, gh.User{Login: "b"},
				[]gh.File{{Path: "api.go"}})
			p.Body = "Some description.\n\nBREAKING CHANGE: removed v1 endpoints"
			return p
		}()},
	}
	rel := Analyze(prs, Options{RootPkg: "root"})
	if len(rel.GlobalChanges) != 2 {
		t.Fatalf("want 2, got %d", len(rel.GlobalChanges))
	}
	if !rel.GlobalChanges[0].Breaking {
		t.Fatalf("breaking change should sort first; order=%v", []string{rel.GlobalChanges[0].Title, rel.GlobalChanges[1].Title})
	}
	if len(rel.BreakingChanges) != 1 {
		t.Fatalf("want 1 breaking, got %d", len(rel.BreakingChanges))
	}
	if rel.BreakingChanges[0].Migration.Summary == "" {
		t.Fatalf("expected migration summary captured")
	}
}

func TestAnalyze_FilterIgnoreBots(t *testing.T) {
	prs := []SourcePR{
		{PR: mkPR(1, "chore(deps): bump x", []string{"dependencies"},
			gh.User{Login: "dependabot[bot]", IsBot: true}, nil)},
		{PR: mkPR(2, "feat: real change", nil, gh.User{Login: "human"}, nil)},
	}
	rel := Analyze(prs, Options{
		RootPkg: "root",
		Filter:  Filter{IgnoreBots: true},
	})
	if len(rel.GlobalChanges) != 1 || rel.GlobalChanges[0].Author.Login != "human" {
		t.Fatalf("expected only human change, got %#v", rel.GlobalChanges)
	}
	if len(rel.Excluded) != 1 || rel.Excluded[0].Reason != "bot" {
		t.Fatalf("expected 1 excluded with reason=bot, got %#v", rel.Excluded)
	}
	if rel.Stats.Excluded != 1 {
		t.Fatalf("stats.excluded: %d", rel.Stats.Excluded)
	}
}

func TestAnalyze_DeterministicOrder(t *testing.T) {
	mkInput := func() []SourcePR {
		return []SourcePR{
			{PR: mkPR(3, "docs: readme", []string{"documentation"}, gh.User{Login: "a"}, nil)},
			{PR: mkPR(1, "feat: a", nil, gh.User{Login: "b"}, nil)},
			{PR: mkPR(2, "fix: b", nil, gh.User{Login: "c"}, nil)},
		}
	}
	a := Analyze(mkInput(), Options{RootPkg: "root"})
	b := Analyze(mkInput(), Options{RootPkg: "root"})
	if len(a.GlobalChanges) != len(b.GlobalChanges) {
		t.Fatalf("len mismatch")
	}
	for i := range a.GlobalChanges {
		if a.GlobalChanges[i].ID != b.GlobalChanges[i].ID {
			t.Fatalf("nondeterministic at %d: %s vs %s", i, a.GlobalChanges[i].ID, b.GlobalChanges[i].ID)
		}
	}
}
