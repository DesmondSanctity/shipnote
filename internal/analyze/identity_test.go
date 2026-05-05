package analyze

import (
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

func TestIdentityMergerResolvesByLogin(t *testing.T) {
	m := NewIdentityMerger([]IdentityRule{
		{
			Canonical: model.Contributor{Login: "maya", Name: "Maya R.", URL: "https://gh/maya"},
			Aliases:   []string{"Maya Rivera", "maya@old.example"},
		},
	})
	got := m.Resolve(model.Contributor{Login: "MAYA", Name: "stale", Changes: 3})
	if got.Login != "maya" || got.Name != "Maya R." || got.Changes != 3 {
		t.Errorf("got %+v", got)
	}
}

func TestIdentityMergerResolvesByAlias(t *testing.T) {
	m := NewIdentityMerger([]IdentityRule{
		{Canonical: model.Contributor{Login: "maya"}, Aliases: []string{"Maya Rivera"}},
	})
	got := m.Resolve(model.Contributor{Login: "", Name: "Maya Rivera", Changes: 1})
	if got.Login != "maya" {
		t.Errorf("alias not resolved: %+v", got)
	}
}

func TestIdentityMergerNoMatchPassesThrough(t *testing.T) {
	m := NewIdentityMerger(nil)
	in := model.Contributor{Login: "guest", Changes: 5}
	if got := m.Resolve(in); got != in {
		t.Errorf("expected identity passthrough, got %+v", got)
	}
}

func TestMergeContributorsAggregatesAndSorts(t *testing.T) {
	m := NewIdentityMerger([]IdentityRule{
		{Canonical: model.Contributor{Login: "maya"}, Aliases: []string{"Maya Rivera"}},
	})
	in := []model.Contributor{
		{Login: "maya", Changes: 2},
		{Name: "Maya Rivera", Changes: 1},
		{Login: "alex", Changes: 4},
		{Login: "Alex", Changes: 1},
	}
	got := MergeContributors(m, in)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].Login != "alex" || got[0].Changes != 5 {
		t.Errorf("alex bucket: %+v", got[0])
	}
	if got[1].Login != "maya" || got[1].Changes != 3 {
		t.Errorf("maya bucket: %+v", got[1])
	}
}

func TestMergeContributorsWithoutMerger(t *testing.T) {
	in := []model.Contributor{
		{Login: "a", Changes: 2},
		{Login: "b", Changes: 5},
	}
	got := MergeContributors(nil, in)
	if got[0].Login != "b" || got[1].Login != "a" {
		t.Errorf("expected sort by Changes desc: %+v", got)
	}
}
