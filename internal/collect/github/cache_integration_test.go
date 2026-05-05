package github_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/collect/github"
)

type brokenTransport struct{}

func (brokenTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network unreachable")
}

func TestGQLCacheHitSkipsNetwork(t *testing.T) {
	fs := newFakeServer(t, happyPathResponse)
	cache := github.NewFileCache(t.TempDir())
	c := github.New(github.Options{Endpoint: fs.URL, Token: "tok", Cache: cache})

	prs, err := c.PRsForCommits(context.Background(), "o", "n", []string{"aaa", "bbb"})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if len(prs) != 2 || prs[0].Number != 42 {
		t.Fatalf("unexpected first result: %+v", prs)
	}

	// Second client points at a broken HTTP transport but shares the
	// same cache directory. The cached response must satisfy the call.
	c2 := github.New(github.Options{
		Endpoint:   fs.URL,
		Token:      "tok",
		HTTPClient: &http.Client{Transport: brokenTransport{}},
		Cache:      cache,
	})
	prs2, err := c2.PRsForCommits(context.Background(), "o", "n", []string{"aaa", "bbb"})
	if err != nil {
		t.Fatalf("second call (cache): %v", err)
	}
	if len(prs2) != 2 || prs2[0].Number != 42 {
		t.Fatalf("unexpected cached result: %+v", prs2)
	}
}

func TestGQLCacheStaleFallbackOnNetworkError(t *testing.T) {
	cacheDir := t.TempDir()
	cache := github.NewFileCache(cacheDir)

	// Seed cache by doing one successful round-trip.
	fs := newFakeServer(t, happyPathResponse)
	c := github.New(github.Options{Endpoint: fs.URL, Cache: cache})
	if _, err := c.PRsForCommits(context.Background(), "o", "n", []string{"aaa", "bbb"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Now point a fresh client at a dead endpoint with the same cache.
	c2 := github.New(github.Options{
		Endpoint:   "http://127.0.0.1:1", // refused
		HTTPClient: &http.Client{Transport: brokenTransport{}},
		Cache:      cache,
	})
	prs, err := c2.PRsForCommits(context.Background(), "o", "n", []string{"aaa", "bbb"})
	if err != nil {
		t.Fatalf("expected stale-cache fallback, got error: %v", err)
	}
	if prs[0].Number != 42 {
		t.Fatalf("stale cache returned wrong data: %+v", prs[0])
	}
}

func TestGQLNoCacheReturnsTransportError(t *testing.T) {
	c := github.New(github.Options{
		Endpoint:   "http://127.0.0.1:1",
		HTTPClient: &http.Client{Transport: brokenTransport{}},
	})
	_, err := c.PRsForCommits(context.Background(), "o", "n", []string{"aaa"})
	if err == nil {
		t.Fatal("expected transport error without cache")
	}
}
