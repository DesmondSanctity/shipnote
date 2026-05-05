package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/collect/github"
)

func TestPRsForCommitsGraphQLError(t *testing.T) {
	t.Parallel()
	fs := newFakeServer(t, `{"errors":[{"message":"boom"}]}`)
	c := github.New(github.Options{Endpoint: fs.URL})
	_, err := c.PRsForCommits(context.Background(), "acme", "sdk", []string{"aaa"})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected graphql error, got %v", err)
	}
}

func TestPRsForCommitsHTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)
	c := github.New(github.Options{Endpoint: srv.URL})
	_, err := c.PRsForCommits(context.Background(), "acme", "sdk", []string{"aaa"})
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected 429 error, got %v", err)
	}
}

func TestPRsForCommitsValidatesArgs(t *testing.T) {
	t.Parallel()
	c := github.New(github.Options{})
	if _, err := c.PRsForCommits(context.Background(), "", "sdk", []string{"x"}); err == nil {
		t.Error("expected error for empty owner")
	}
	got, err := c.PRsForCommits(context.Background(), "a", "b", nil)
	if err != nil || got != nil {
		t.Errorf("empty shas should be (nil, nil); got %v %v", got, err)
	}
}
