package analyze

import (
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

func TestCategorizeFromLabel(t *testing.T) {
	cases := []struct {
		labels []string
		want   model.ChangeType
	}{
		{[]string{"bug"}, model.ChangeTypeFix},
		{[]string{"type:feat"}, model.ChangeTypeFeat},
		{[]string{"kind/bug"}, model.ChangeTypeFix},
		{[]string{"Performance"}, model.ChangeTypePerf},
		{[]string{"unrelated", "docs"}, model.ChangeTypeDocs},
	}
	for _, tc := range cases {
		got := Categorize(Input{Labels: tc.labels})
		if got.Type != tc.want {
			t.Errorf("labels=%v: got %s, want %s", tc.labels, got.Type, tc.want)
		}
		if got.Source != "label" {
			t.Errorf("labels=%v: source=%s, want label", tc.labels, got.Source)
		}
	}
}

func TestCategorizeFromChangelogBlock(t *testing.T) {
	in := Input{PRBody: "fixes things\n\nchangelog: fix\n"}
	got := Categorize(in)
	if got.Type != model.ChangeTypeFix || got.Source != "changelog-block" {
		t.Fatalf("got %+v", got)
	}
}

func TestCategorizeFromConventionalCommit(t *testing.T) {
	cases := []struct {
		title string
		want  model.ChangeType
	}{
		{"feat: add login", model.ChangeTypeFeat},
		{"fix(api): null pointer", model.ChangeTypeFix},
		{"feat!: drop legacy api", model.ChangeTypeFeat},
		{"perf(core)!: rewrite hot path", model.ChangeTypePerf},
		{"chore: bump deps", model.ChangeTypeChore},
	}
	for _, tc := range cases {
		got := Categorize(Input{PRTitle: tc.title})
		if got.Type != tc.want {
			t.Errorf("%q -> %s, want %s", tc.title, got.Type, tc.want)
		}
		if got.Source != "cc-prefix" {
			t.Errorf("%q source=%s", tc.title, got.Source)
		}
	}
}

func TestCategorizeFromHeuristic(t *testing.T) {
	got := Categorize(Input{FilesPaths: []string{"docs/intro.md", "README.md"}})
	if got.Type != model.ChangeTypeDocs || got.Source != "heuristic" {
		t.Errorf("docs heuristic: %+v", got)
	}
	got = Categorize(Input{FilesPaths: []string{"foo_test.go", "bar_test.go"}})
	if got.Type != model.ChangeTypeTest || got.Source != "heuristic" {
		t.Errorf("test heuristic: %+v", got)
	}
}

func TestCategorizeDefault(t *testing.T) {
	got := Categorize(Input{PRTitle: "random unstructured title"})
	if got.Type != model.ChangeTypeOther || got.Source != "default" {
		t.Errorf("default: %+v", got)
	}
}

func TestCategorizePrecedenceLabelWinsOverCC(t *testing.T) {
	got := Categorize(Input{
		Labels:  []string{"docs"},
		PRTitle: "feat: would-be-feature",
	})
	if got.Type != model.ChangeTypeDocs {
		t.Errorf("expected label to beat CC: %+v", got)
	}
}
