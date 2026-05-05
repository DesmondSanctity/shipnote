package github_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DesmondSanctity/shipnote/internal/collect/github"
)

// pagedServer responds to each request by reflecting the SHAs back as
// PR numbers (one per c<i> alias seen in the query). It tracks
// concurrent in-flight requests so tests can assert the bound.
type pagedServer struct {
	*httptest.Server
	inflight  atomic.Int32
	maxSeen   atomic.Int32
	totalCall atomic.Int32
}

func newPagedServer(t *testing.T) *pagedServer {
	t.Helper()
	ps := &pagedServer{}
	ps.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ps.totalCall.Add(1)
		now := ps.inflight.Add(1)
		for {
			peak := ps.maxSeen.Load()
			if now <= peak || ps.maxSeen.CompareAndSwap(peak, now) {
				break
			}
		}
		// Hold briefly so concurrent calls overlap.
		time.Sleep(20 * time.Millisecond)
		defer ps.inflight.Add(-1)

		body, _ := io.ReadAll(r.Body)
		// Count occurrences of "c<i>:" to know the page size.
		// Naive but enough for fixtures.
		count := 0
		for i := 0; i < 1000; i++ {
			needle := fmt.Sprintf("c%d:", i)
			if !contains(string(body), needle) {
				count = i
				break
			}
		}
		// Build a response with `count` empty PR slots; gives us
		// deterministic length without exercising decode logic.
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"repository":{`))
		for i := 0; i < count; i++ {
			if i > 0 {
				_, _ = w.Write([]byte(","))
			}
			_, _ = fmt.Fprintf(w, `"c%d":{"oid":"x","associatedPullRequests":{"nodes":[]}}`, i)
		}
		_, _ = w.Write([]byte(`}}}`))
	}))
	t.Cleanup(ps.Close)
	return ps
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestPRsForCommitsPagedSplitsAndBoundsConcurrency(t *testing.T) {
	ps := newPagedServer(t)
	c := github.New(github.Options{Endpoint: ps.URL, Token: "tok"})

	shas := make([]string, 25)
	for i := range shas {
		shas[i] = fmt.Sprintf("sha%02d", i)
	}

	prs, err := c.PRsForCommitsPaged(context.Background(), "o", "n", shas, github.PageOptions{
		PageSize:    10,
		Concurrency: 3,
	})
	if err != nil {
		t.Fatalf("paged: %v", err)
	}
	if len(prs) != len(shas) {
		t.Fatalf("len = %d, want %d", len(prs), len(shas))
	}
	if got := ps.totalCall.Load(); got != 3 {
		t.Errorf("expected 3 page calls (10+10+5), got %d", got)
	}
	if peak := ps.maxSeen.Load(); peak > 3 {
		t.Errorf("concurrency bound exceeded: max in-flight = %d", peak)
	}
	if peak := ps.maxSeen.Load(); peak < 2 {
		t.Errorf("expected at least 2 concurrent calls, got %d", peak)
	}
}

func TestPRsForCommitsPagedAnonymousCapsConcurrency(t *testing.T) {
	ps := newPagedServer(t)
	c := github.New(github.Options{Endpoint: ps.URL}) // no token = anonymous

	shas := make([]string, 30)
	for i := range shas {
		shas[i] = fmt.Sprintf("sha%02d", i)
	}
	if _, err := c.PRsForCommitsPaged(context.Background(), "o", "n", shas, github.PageOptions{
		PageSize:    5,
		Concurrency: 16, // requested high; should be capped
	}); err != nil {
		t.Fatalf("paged: %v", err)
	}
	if peak := ps.maxSeen.Load(); peak > 2 {
		t.Errorf("anonymous concurrency cap broken: max = %d", peak)
	}
}

func TestPRsForCommitsPagedEmpty(t *testing.T) {
	c := github.New(github.Options{Endpoint: "http://unused"})
	prs, err := c.PRsForCommitsPaged(context.Background(), "o", "n", nil, github.DefaultPageOptions())
	if err != nil {
		t.Fatal(err)
	}
	if prs != nil {
		t.Fatalf("expected nil, got %v", prs)
	}
}

func TestPRsForCommitsPagedValidatesArgs(t *testing.T) {
	c := github.New(github.Options{Endpoint: "http://unused"})
	if _, err := c.PRsForCommitsPaged(context.Background(), "", "n", []string{"x"}, github.DefaultPageOptions()); err == nil {
		t.Fatal("expected error for empty owner")
	}
}
