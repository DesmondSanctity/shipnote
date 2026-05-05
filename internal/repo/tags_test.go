package repo_test

import (
	"context"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/repo"
)

func TestTagsAndLatestTag(t *testing.T) {
	t.Parallel()
	dir := fixture(t)
	g, err := repo.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	tags, err := g.Tags(context.Background(), "v*")
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "v0.1.0" {
		t.Fatalf("unexpected tags: %+v", tags)
	}
	if tags[0].SHA == "" || tags[0].Date == "" {
		t.Fatalf("tag missing fields: %+v", tags[0])
	}

	latest, ok, err := g.LatestTag(context.Background(), "v*")
	if err != nil || !ok {
		t.Fatalf("latest tag: ok=%v err=%v", ok, err)
	}
	if latest.Name != "v0.1.0" {
		t.Fatalf("unexpected latest: %+v", latest)
	}

	none, ok, err := g.LatestTag(context.Background(), "z*")
	if err != nil {
		t.Fatalf("err on no-match: %v", err)
	}
	if ok || none.Name != "" {
		t.Fatalf("expected no match for z*, got %+v", none)
	}
}
