package site

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

func sampleRelease(tag, date string) model.Release {
	pub := date
	return model.Release{
		SchemaVersion: model.SchemaVersion,
		ID:            "rel-" + tag,
		Metadata:      model.Metadata{GeneratedAt: date, Generator: model.Generator{Name: "shipnote", Version: "test"}},
		Repo:          model.Repo{Provider: "github", Owner: "acme", Name: "widgets", URL: "https://github.com/acme/widgets"},
		Range: model.Range{
			From: model.RangePoint{Ref: "v0.0.1", SHA: "aaaa", Date: "2025-01-01"},
			To:   model.RangePoint{Ref: tag, SHA: "bbbb", Date: date},
		},
		Release: model.ReleaseHeader{State: model.ReleaseStateReleased, Tag: tag, Name: tag, PublishedAt: &pub},
		Summary: model.Summary{
			Headline:    "Test release " + tag,
			AISummaries: model.AISummaries{Customer: &model.AISummary{Text: "Customer-friendly recap.", Model: "test", PromptVersion: "v1"}},
		},
		Stats: model.Stats{Features: 1, Fixes: 1, Breaking: 0},
		GlobalChanges: []model.Change{
			{ID: "c1", Type: model.ChangeTypeFeat, Title: "Add a thing"},
			{ID: "c2", Type: model.ChangeTypeFix, Title: "Fix a thing"},
		},
		Contributors: []model.Contributor{{Login: "octocat", Changes: 2, URL: "https://github.com/octocat"}},
	}
}

func TestBuildIdempotent(t *testing.T) {
	dir := t.TempDir()
	rel := sampleRelease("v0.1.0", "2025-02-01")
	opts := Options{Root: dir, SiteURL: "https://example.com/cl", Title: "Acme"}
	if err := Build(opts, rel); err != nil {
		t.Fatalf("first build: %v", err)
	}
	first := snapshot(t, dir)
	if err := Build(opts, rel); err != nil {
		t.Fatalf("second build: %v", err)
	}
	second := snapshot(t, dir)
	for path, body := range first {
		if !bytes.Equal(body, second[path]) {
			t.Errorf("not idempotent: %s differs", path)
		}
	}
}

func TestBuildAddsRelease(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Root: dir, SiteURL: "https://example.com/cl"}
	if err := Build(opts, sampleRelease("v0.1.0", "2025-01-15")); err != nil {
		t.Fatal(err)
	}
	if err := Build(opts, sampleRelease("v0.2.0", "2025-02-15")); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"index.html", "v0.1.0/index.html", "v0.2.0/index.html",
		"feed.json", "feed.xml", "releases.json",
		"assets/site.css", "assets/widget.js",
	}
	for _, p := range want {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}
	home, _ := os.ReadFile(filepath.Join(dir, "index.html"))
	if !strings.Contains(string(home), "v0.2.0") {
		t.Errorf("home page missing latest tag")
	}
	feedRaw, _ := os.ReadFile(filepath.Join(dir, "feed.json"))
	var feed map[string]any
	if err := json.Unmarshal(feedRaw, &feed); err != nil {
		t.Fatalf("feed json: %v", err)
	}
	items, _ := feed["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	first := items[0].(map[string]any)
	if _, ok := first["_shipnote"]; !ok {
		t.Errorf("first item missing _shipnote extension")
	}
}

func TestUpsertReplacesByTag(t *testing.T) {
	idx := &siteIndex{}
	idx.upsert(sampleRelease("v0.1.0", "2025-01-01"))
	idx.upsert(sampleRelease("v0.1.0", "2025-01-02"))
	if got := len(idx.Entries); got != 1 {
		t.Fatalf("want 1 entry, got %d", got)
	}
	if got := idx.Entries[0].Date; got != "2025-01-02" {
		t.Errorf("want updated date, got %q", got)
	}
}

func snapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		out[rel] = body
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return out
}
