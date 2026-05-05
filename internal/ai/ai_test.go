package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/ai"
	"github.com/DesmondSanctity/shipnote/internal/model"
)

// fakeOpenAI returns a server that mimics the OpenAI Chat Completions
// endpoint and emits a deterministic response per request. /models is
// answered with 200 so Available() succeeds.
func fakeOpenAI(t *testing.T, calls *int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/chat/completions":
			if calls != nil {
				*calls++
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"model": "stub-model",
				"choices": []map[string]any{{
					"message": map[string]any{"role": "assistant", "content": "Stub summary."},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestOpenAICompatible_Available(t *testing.T) {
	t.Parallel()
	srv := fakeOpenAI(t, nil)
	defer srv.Close()
	c := ai.NewOpenAICompatible(ai.OpenAICompatibleConfig{
		Name: "stub", BaseURL: srv.URL, Model: "stub-model",
	})
	ok, err := c.Available(context.Background())
	if err != nil || !ok {
		t.Fatalf("Available: ok=%v err=%v", ok, err)
	}
}

func TestEnrich_PopulatesAudiencesAndCaches(t *testing.T) {
	t.Parallel()
	calls := 0
	srv := fakeOpenAI(t, &calls)
	defer srv.Close()
	prov := ai.NewOpenAICompatible(ai.OpenAICompatibleConfig{
		Name: "stub", BaseURL: srv.URL, Model: "stub-model",
	})
	cacheDir := t.TempDir()
	cache := &ai.Cache{Dir: cacheDir}

	release := func() *model.Release {
		return &model.Release{
			Repo: model.Repo{Owner: "Acme", Name: "widgets"},
			GlobalChanges: []model.Change{{
				ID: "github-prs:7", Type: model.ChangeTypeFix,
				Title: "fix crash on empty input",
			}},
		}
	}

	rel := release()
	if err := ai.Enrich(context.Background(), rel, ai.Options{
		Provider: prov, Cache: cache,
	}); err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if rel.Summary.AISummaries.Developer == nil {
		t.Fatalf("expected release-level developer summary")
	}
	if !strings.Contains(rel.Summary.AISummaries.Developer.Text, "Stub summary") {
		t.Fatalf("unexpected text: %q", rel.Summary.AISummaries.Developer.Text)
	}
	if rel.GlobalChanges[0].AISummaries.Customer == nil {
		t.Fatalf("expected per-change customer summary")
	}
	if rel.Summary.AISummaries.Developer.PromptVersion != ai.PromptVersion {
		t.Fatalf("expected PromptVersion %s, got %q", ai.PromptVersion, rel.Summary.AISummaries.Developer.PromptVersion)
	}

	// 4 scopes (1 release + 1 change) x 3 audiences = 6 calls on cold run.
	if calls != 6 {
		t.Fatalf("expected 6 cold calls, got %d", calls)
	}

	// Second pass with the same cache must hit zero new HTTP calls.
	calls = 0
	rel2 := release()
	if err := ai.Enrich(context.Background(), rel2, ai.Options{
		Provider: prov, Cache: cache,
	}); err != nil {
		t.Fatalf("Enrich (warm): %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected 0 warm calls, got %d", calls)
	}
	if rel2.Summary.AISummaries.Developer == nil ||
		rel2.Summary.AISummaries.Developer.Text != rel.Summary.AISummaries.Developer.Text {
		t.Fatalf("warm pass produced different output")
	}
}

func TestEnrich_NilProviderIsNoOp(t *testing.T) {
	t.Parallel()
	rel := &model.Release{GlobalChanges: []model.Change{{ID: "x", Title: "y"}}}
	if err := ai.Enrich(context.Background(), rel, ai.Options{}); err != nil {
		t.Fatalf("Enrich nil provider: %v", err)
	}
	if rel.Summary.AISummaries.Developer != nil {
		t.Fatalf("expected no summary when provider nil")
	}
}
