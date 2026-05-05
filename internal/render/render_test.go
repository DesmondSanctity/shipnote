package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

func sampleRelease() model.Release {
	return model.Release{
		SchemaVersion: model.SchemaVersion,
		Metadata:      model.Metadata{Generator: model.Generator{Name: "shipnote", Version: "0.1.0"}},
		Release:       model.ReleaseHeader{Tag: "v1.2.0", State: model.ReleaseStateReleased},
		Range:         model.Range{To: model.RangePoint{Date: "2025-04-01T00:00:00Z"}},
		GlobalChanges: []model.Change{
			{
				ID:    "github-prs:7",
				Type:  model.ChangeTypeFix,
				Title: "fix: avoid panic on empty input",
				Source: model.ChangeSource{
					Kind: model.SourceKindPR,
					PR:   &model.PRSource{Number: 7, URL: "https://example.com/pr/7"},
				},
				Author: model.Contributor{Login: "alice", URL: "https://example.com/alice"},
			},
			{
				ID:       "github-prs:8",
				Type:     model.ChangeTypeFeat,
				Breaking: true,
				Title:    "feat: drop legacy api",
				Source: model.ChangeSource{
					Kind: model.SourceKindPR,
					PR:   &model.PRSource{Number: 8, URL: "https://example.com/pr/8"},
				},
				Author:    model.Contributor{Login: "bob"},
				Migration: &model.Migration{Summary: "remove use of /v1"},
			},
		},
		BreakingChanges: []model.BreakingChange{{ChangeID: "github-prs:8", Summary: "feat: drop legacy api"}},
		Contributors: []model.Contributor{
			{Login: "alice", URL: "https://example.com/alice", Changes: 1},
			{Login: "bob", Changes: 1},
		},
	}
}

func TestMarkdown_HeaderAndStickyMarkers(t *testing.T) {
	out := Markdown(sampleRelease())
	if !strings.HasPrefix(out, StickyStart+"\n") {
		t.Fatalf("missing sticky-start prefix: %q", out[:40])
	}
	if !strings.HasSuffix(out, StickyEnd+"\n") {
		t.Fatalf("missing sticky-end suffix")
	}
	if !strings.Contains(out, "## [v1.2.0] - 2025-04-01") {
		t.Fatalf("missing header line, got:\n%s", out)
	}
}

func TestMarkdown_BreakingFloatsAndMigration(t *testing.T) {
	out := Markdown(sampleRelease())
	breakIdx := strings.Index(out, "### Breaking Changes")
	fixIdx := strings.Index(out, "### Fixed")
	if breakIdx < 0 || fixIdx < 0 || breakIdx > fixIdx {
		t.Fatalf("breaking section must precede fixed; got break=%d fix=%d\n%s", breakIdx, fixIdx, out)
	}
	if !strings.Contains(out, "_Migration:_ remove use of /v1") {
		t.Fatalf("migration line missing")
	}
	if !strings.Contains(out, "([#8](https://example.com/pr/8))") {
		t.Fatalf("PR link missing")
	}
	if !strings.Contains(out, "by [@alice](https://example.com/alice)") {
		t.Fatalf("author link missing")
	}
}

func TestMarkdown_ContributorsSorted(t *testing.T) {
	rel := sampleRelease()
	rel.Contributors = []model.Contributor{
		{Login: "zoe", Changes: 1},
		{Login: "alice", Changes: 3},
		{Login: "bob", Changes: 3},
	}
	out := Markdown(rel)
	i := strings.Index(out, "### Contributors")
	if i < 0 {
		t.Fatalf("missing contributors section")
	}
	tail := out[i:]
	posAlice := strings.Index(tail, "@alice")
	posBob := strings.Index(tail, "@bob")
	posZoe := strings.Index(tail, "@zoe")
	if posAlice >= posBob || posBob >= posZoe {
		t.Fatalf("contributors not sorted by changes desc, login asc:\n%s", tail)
	}
}

func TestMarkdown_Empty(t *testing.T) {
	out := Markdown(model.Release{Release: model.ReleaseHeader{}})
	if !strings.Contains(out, "## [Unreleased]") {
		t.Fatalf("expected Unreleased header, got:\n%s", out)
	}
	if strings.Contains(out, "### Contributors") {
		t.Fatalf("contributors section should not appear when empty")
	}
}

func TestMarkdown_Deterministic(t *testing.T) {
	a := Markdown(sampleRelease())
	b := Markdown(sampleRelease())
	if a != b {
		t.Fatalf("non-deterministic output")
	}
}

func TestJSON_RoundTrip(t *testing.T) {
	data, err := JSON(sampleRelease())
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var back model.Release
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.SchemaVersion != model.SchemaVersion {
		t.Fatalf("schema version: %s", back.SchemaVersion)
	}
	if len(back.GlobalChanges) != 2 {
		t.Fatalf("changes lost in round-trip: %d", len(back.GlobalChanges))
	}
}

func TestJSON_Deterministic(t *testing.T) {
	a, err := JSON(sampleRelease())
	if err != nil {
		t.Fatalf("a: %v", err)
	}
	b, err := JSON(sampleRelease())
	if err != nil {
		t.Fatalf("b: %v", err)
	}
	if string(a) != string(b) {
		t.Fatalf("non-deterministic JSON output")
	}
}

func TestJSON_NoHTMLEscape(t *testing.T) {
	rel := sampleRelease()
	rel.GlobalChanges[0].Title = "fix <script> & ampersand"
	data, err := JSON(rel)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if !strings.Contains(string(data), "<script>") {
		t.Fatalf("expected HTML chars un-escaped, got:\n%s", data)
	}
}
