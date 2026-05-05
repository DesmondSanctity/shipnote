package github

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := NewFileCache(dir)

	if _, err := c.Get("nope"); err != ErrCacheMiss {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}

	body := []byte(`{"hello":"world"}`)
	if err := c.Put("k1", body); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := c.Get("k1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("body mismatch: %s", got)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	if len(entries[0].Name()) != 64 {
		t.Fatalf("expected 64-char hex name, got %q", entries[0].Name())
	}
}

func TestFileCacheReplaceAtomic(t *testing.T) {
	dir := t.TempDir()
	c := NewFileCache(dir)
	if err := c.Put("k", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := c.Put("k", []byte("v2-longer")); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get("k")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v2-longer" {
		t.Fatalf("expected v2-longer, got %s", got)
	}
}

func TestHashKeyIdempotent(t *testing.T) {
	h := HashKey("query+vars")
	if HashKey(h) != h {
		t.Fatal("HashKey not idempotent on already-hashed input")
	}
}

func TestFileCacheMissingDir(t *testing.T) {
	c := NewFileCache(filepath.Join(t.TempDir(), "nested", "missing"))
	if _, err := c.Get("k"); err != ErrCacheMiss {
		t.Fatalf("expected miss, got %v", err)
	}
}
